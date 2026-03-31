<script setup>
import { ref } from "vue"

const props = defineProps({
  node: {
    type: Object,
    required: true
  },
  level: {
    type: Number,
    default: 0
  }
})

const collapsed = ref(false)
const showReplyEditor = ref(false)
const replyText = ref("")
const emit = defineEmits(["publish-reply"])

function toggleReplies() {
  collapsed.value = !collapsed.value
}

function toggleReplyEditor() {
  showReplyEditor.value = !showReplyEditor.value
}

function publishReply() {
  const text = replyText.value.trim()
  if (!text) return
  emit("publish-reply", { parentId: props.node.id, content: text })
  replyText.value = ""
  showReplyEditor.value = false
}
</script>

<template>
  <li class="thread-item" :style="{ marginLeft: `${Math.min(props.level * 14, 56)}px` }">
    <div class="meta">
      <span class="author">{{ props.node.author || "匿名用户" }}</span>
      <span class="time">{{ props.node.createdAt || "未知时间" }}</span>
    </div>
    <p class="content">{{ props.node.content || "（无内容）" }}</p>

    <div class="actions">
      <button type="button" class="reply-action-btn" @click="toggleReplyEditor">
        {{ showReplyEditor ? "取消回复" : "回复" }}
      </button>
      <button
        v-if="props.node.replies?.length"
        type="button"
        class="reply-toggle-btn"
        @click="toggleReplies"
      >
        {{ collapsed ? "展开回复" : "收起回复" }}（{{ props.node.replies.length }}）
      </button>
    </div>

    <div v-if="showReplyEditor" class="reply-editor">
      <textarea
        v-model="replyText"
        class="reply-input"
        rows="2"
        placeholder="输入回复内容..."
      ></textarea>
      <div class="reply-editor-actions">
        <button type="button" class="publish-btn" @click="publishReply">发布</button>
      </div>
    </div>

    <div v-if="props.node.replies?.length" class="reply-wrap">
      <transition name="collapse">
        <ul v-show="!collapsed" class="reply-list">
          <CommentThread
            v-for="reply in props.node.replies"
            :key="reply.id"
            :node="reply"
            :level="props.level + 1"
            @publish-reply="emit('publish-reply', $event)"
          />
        </ul>
      </transition>
    </div>
  </li>
</template>

<style scoped>
.thread-item {
  border: 1px solid rgba(148, 163, 184, 0.16);
  background: rgba(2, 6, 23, 0.5);
  border-radius: 10px;
  padding: 10px 12px;
}

.meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.author {
  color: #e2e8f0;
  font-size: 13px;
  font-weight: 700;
}

.time {
  color: #94a3b8;
  font-size: 12px;
}

.content {
  margin: 0;
  color: #cbd5e1;
  line-height: 1.6;
  font-size: 14px;
}

.reply-wrap {
  margin-top: 10px;
}

.actions {
  margin-top: 10px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.reply-action-btn {
  border: 1px solid rgba(148, 163, 184, 0.22);
  background: rgba(51, 65, 85, 0.72);
  color: #e2e8f0;
  border-radius: 8px;
  padding: 5px 10px;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
}

.reply-action-btn:hover {
  background: rgba(71, 85, 105, 0.86);
}

.reply-toggle-btn {
  border: 1px solid rgba(148, 163, 184, 0.2);
  background: rgba(30, 41, 59, 0.6);
  color: #cbd5e1;
  border-radius: 8px;
  padding: 5px 10px;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
}

.reply-toggle-btn:hover {
  background: rgba(51, 65, 85, 0.86);
}

.reply-editor {
  margin-top: 8px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background: rgba(15, 23, 42, 0.45);
  border-radius: 8px;
  padding: 8px;
}

.reply-input {
  width: 100%;
  box-sizing: border-box;
  resize: vertical;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 8px;
  background: rgba(15, 23, 42, 0.75);
  color: #e2e8f0;
  padding: 8px;
  outline: none;
}

.reply-input:focus {
  border-color: rgba(96, 165, 250, 0.45);
}

.reply-editor-actions {
  margin-top: 8px;
  display: flex;
  justify-content: flex-end;
}

.publish-btn {
  border: 1px solid rgba(96, 165, 250, 0.35);
  background: rgba(37, 99, 235, 0.3);
  color: #dbeafe;
  border-radius: 8px;
  padding: 5px 10px;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
}

.publish-btn:hover {
  background: rgba(37, 99, 235, 0.5);
}

.reply-list {
  margin: 10px 0 0;
  padding: 0 0 0 14px;
  list-style: none;
  border-left: 2px solid rgba(148, 163, 184, 0.22);
  display: grid;
  gap: 8px;
}

.collapse-enter-active,
.collapse-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.collapse-enter-from,
.collapse-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
