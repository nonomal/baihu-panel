<script setup lang="ts">
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Plus, Trash2, TestTube, Pencil, Bell } from 'lucide-vue-next'
import type { NotifyChannel, ChannelType } from '@/api'

defineProps<{
  channels: NotifyChannel[]
  channelTypes: ChannelType[]
}>()

const emit = defineEmits<{
  add: []
  edit: [channel: NotifyChannel]
  delete: [id: string]
  test: [channel: NotifyChannel]
}>()

function getChannelTypeName(type: string, channelTypes: ChannelType[]): string {
  const found = channelTypes.find(t => t.type === type)
  return found ? found.label : type
}
</script>

<template>
  <Card>
    <CardHeader>
      <div class="flex items-center justify-between">
        <div>
          <CardTitle>通知渠道</CardTitle>
          <CardDescription>管理消息推送渠道配置</CardDescription>
        </div>
        <Button size="sm" @click="emit('add')">
          <Plus class="w-4 h-4 mr-1" />
          添加渠道
        </Button>
      </div>
    </CardHeader>
    <CardContent>
      <div v-if="channels.length === 0" class="text-center py-12 text-muted-foreground">
        <Bell class="w-12 h-12 mx-auto mb-3 opacity-30" />
        <p class="text-sm">暂无通知渠道</p>
        <p class="text-xs mt-1">点击"添加渠道"开始配置</p>
      </div>
      <div v-else class="space-y-3">
        <div
          v-for="ch in channels"
          :key="ch.id"
          class="flex items-center justify-between p-3 rounded-lg border bg-card hover:bg-accent/30 transition-colors"
        >
          <div class="flex items-center gap-3 min-w-0 flex-1">
            <div class="flex flex-col min-w-0">
              <div class="flex items-center gap-2">
                <span class="font-medium text-sm truncate">{{ ch.name }}</span>
                <Badge variant="secondary" class="text-[10px] shrink-0">{{ getChannelTypeName(ch.type, channelTypes) }}</Badge>
                <Badge
                  :class="ch.enabled ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20' : 'bg-zinc-500/10 text-zinc-500 border-zinc-500/20'"
                  variant="secondary"
                  class="text-[10px] shrink-0"
                >
                  {{ ch.enabled ? '启用' : '禁用' }}
                </Badge>
              </div>
            </div>
          </div>
          <div class="flex items-center gap-1 shrink-0">
            <Button variant="ghost" size="icon" class="h-7 w-7" @click="emit('test', ch)" title="测试发送">
              <TestTube class="w-3.5 h-3.5" />
            </Button>
            <Button variant="ghost" size="icon" class="h-7 w-7" @click="emit('edit', ch)" title="编辑">
              <Pencil class="w-3.5 h-3.5" />
            </Button>
            <Button variant="ghost" size="icon" class="h-7 w-7 text-destructive hover:text-destructive" @click="emit('delete', ch.id)" title="删除">
              <Trash2 class="w-3.5 h-3.5" />
            </Button>
          </div>
        </div>
      </div>
    </CardContent>
  </Card>
</template>
