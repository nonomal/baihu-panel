package router

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/engigu/baihu-panel/internal/static"

	"github.com/gin-gonic/gin"
)

func mustSubFS(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return sub
}

// cacheControl 返回设置 Cache-Control header 的中间件
func cacheControl(value string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", value)
		c.Next()
	}
}

func initStaticRoutes(root *gin.RouterGroup) {
	staticFS := static.GetFS()
	if staticFS == nil {
		return
	}

	assetsFS := mustSubFS(staticFS, "assets")
	
	// 静态资源路由：专门处理 assets 目录
	root.GET("/assets/*filepath", cacheControl("public, max-age=31536000, immutable"), func(ctx *gin.Context) {
		path := strings.TrimPrefix(ctx.Param("filepath"), "/")
		if path == "" {
			ctx.Status(404)
			return
		}

		// 智能选择资源：优先寻找 .gz 版本，即便请求的是原文件名 (如 typescript.js)
		gzPath := path + ".gz"
		isGzipSupported := strings.Contains(ctx.GetHeader("Accept-Encoding"), "gzip")

		// 检查 .gz 文件是否存在
		if gzFile, err := assetsFS.Open(gzPath); err == nil {
			defer gzFile.Close()
			
			contentType := mime.TypeByExtension(filepath.Ext(path))
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			ctx.Header("Content-Type", contentType)

			if isGzipSupported {
				// 1. 客户端支持 Gzip: 直接发送压缩后的数据 (最优解)
				ctx.Header("Content-Encoding", "gzip")
				io.Copy(ctx.Writer, gzFile)
			} else {
				// 2. 客户端不支持 Gzip: 现场解压给它 (兼容性退路)
				gr, _ := gzip.NewReader(gzFile)
				defer gr.Close()
				io.Copy(ctx.Writer, gr)
			}
			return
		}

		// 3. 如果连 .gz 都没有，最后尝试返回原文件（比如图片等本身不适合压缩的资源）
		if file, err := assetsFS.Open(path); err == nil {
			defer file.Close()
			http.FileServer(http.FS(assetsFS)).ServeHTTP(ctx.Writer, ctx.Request)
			return
		}

		ctx.Status(404)
	})

	// logo.svg 短缓存实现
	root.GET("/logo.svg", func(ctx *gin.Context) {
		serveSingleFile(ctx, "logo.svg", "image/svg+xml", "public, max-age=86400")
	})
}

// serveSingleFile 处理单个静态文件的逻辑（支持自动解压）
func serveSingleFile(ctx *gin.Context, filename string, contentType string, cache string) {
	if cache != "" {
		ctx.Header("Cache-Control", cache)
	}
	ctx.Header("Content-Type", contentType)

	// 尝试寻找压缩版
	if gzData, err := static.ReadFile(filename + ".gz"); err == nil {
		if strings.Contains(ctx.GetHeader("Accept-Encoding"), "gzip") {
			ctx.Header("Content-Encoding", "gzip")
			ctx.Data(200, contentType, gzData)
		} else {
			gr, _ := gzip.NewReader(bytes.NewReader(gzData))
			defer gr.Close()
			io.Copy(ctx.Writer, gr)
		}
		return
	}

	// 尝试原文件
	if data, err := static.ReadFile(filename); err == nil {
		ctx.Data(200, contentType, data)
		return
	}
	
	ctx.Status(404)
}

// serveSPA 注入配置并返回 index.html 给前端渲染
func serveSPA(ctx *gin.Context, urlPrefix string, status int) {
	var data []byte
	
	// 尝试解压 index.html.gz (因为我们需要修改其内容，不能直接发 gz)
	if gzData, err := static.ReadFile("index.html.gz"); err == nil {
		gr, _ := gzip.NewReader(bytes.NewReader(gzData))
		data, _ = io.ReadAll(gr)
		gr.Close()
	} else {
		// 回退到普通 index.html
		data, _ = static.ReadFile("index.html")
	}

	if data == nil {
		// ... 保持原有 fallback 逻辑 ...
		serveFallback(ctx, urlPrefix, status)
		return
	}

	html := string(data)
	baseHref := urlPrefix + "/"
	if urlPrefix == "" { baseHref = "/" }
	html = strings.Replace(html, "<head>", "<head>\n    <base href=\""+baseHref+"\">", 1)
	configScript := `<script>window.__BASE_URL__ = "` + urlPrefix + `"; window.__API_VERSION__ = "/api/v1";</script>`
	html = strings.Replace(html, "</head>", configScript+"</head>", 1)

	ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ctx.Data(status, "text/html; charset=utf-8", []byte(html))
}

func serveFallback(ctx *gin.Context, urlPrefix string, status int) {
	path := ctx.Request.URL.Path
	if strings.HasSuffix(path, "/404") {
		ctx.Data(status, "text/html; charset=utf-8", []byte("<!DOCTYPE html><html>..."))
		ctx.Abort()
		return
	}
	// ... 原有逻辑 ...
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Data(status, "text/html", []byte("Not Found"))
	ctx.Abort()
}
