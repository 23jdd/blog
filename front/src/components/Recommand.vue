<script setup>
import { computed, ref, watch } from "vue"
import CommentThread from "./CommentThread.vue"
import { createArticleComment, fetchArticleComments } from "@/api/blogApi.js"

const props = defineProps({
  title: {
    type: String,
    default: "评论区"
  },
  articleId: {
    type: [Number, String],
    default: null
  },
  comments: {
    type: Array,
    default: () => [
      {
        id: 1,
        author: "Alice",
        content: "这篇文章的结构很清晰，感谢分享。",
        createdAt: "2026-03-26 10:12",
        replies: [
          {
            id: "1-1",
            author: "Long",
            content: "感谢反馈，我会继续补充实战案例。",
            createdAt: "2026-03-26 10:30",
            replies:[
                {
                id :"1-1-1",       
                author: "Long",
                content: "感谢",
                createdAt: "2026-03-26 10:30",
                }
            ]
          }
        ]
      },
      {
        id: 2,
        author: "Bob",
        content: "希望后续能补充更多实战案例。",
        createdAt: "2026-03-26 11:03",
        replies: []
      }
    ]
  }
})

const collapsed = ref(false)
const localComments = ref([])
const newCommentText = ref("")
const loading = ref(false)
const error = ref("")
const posting = ref(false)
const emit = defineEmits(["toggle"])

watch(
  () => props.comments,
  (next) => {
    localComments.value = JSON.parse(JSON.stringify(next || []))
  },
  { immediate: true, deep: true }
)

const commentCount = computed(() => localComments.value.length)

function mapReviewToNode(r) {
  return {
    id: r.id,
    parentId: r.parentID ?? 0,
    author: r.authorName || `用户#${r.authorID ?? "匿名"}`,
    content: r.content || "",
    createdAt: r.createTime || "",
    replies: []
  }
}

function buildCommentTree(reviews) {
  const byId = new Map()
  const nodes = reviews.map((r) => mapReviewToNode(r))
  for (const n of nodes) byId.set(String(n.id), n)
  const roots = []
  for (const n of nodes) {
    const pid = Number(n.parentId || 0)
    if (!pid) {
      roots.push(n)
      continue
    }
    const parent = byId.get(String(pid))
    if (parent) parent.replies.unshift(n)
    else roots.push(n)
  }
  roots.sort((a, b) => Number(b.id) - Number(a.id))
  return roots
}

async function loadComments() {
  const id = Number(props.articleId)
  if (!id || Number.isNaN(id)) return
  loading.value = true
  error.value = ""
  try {
    const rows = await fetchArticleComments(id, 1, 500)
    localComments.value = buildCommentTree(rows)
  } catch (e) {
    error.value = e?.message || "加载评论失败"
  } finally {
    loading.value = false
  }
}

watch(
  () => props.articleId,
  () => {
    if (props.articleId != null) {
      loadComments()
    }
  },
  { immediate: true }
)

function toggleComments() {
  collapsed.value = !collapsed.value
  emit("toggle", { collapsed: collapsed.value })
}

async function publishComment() {
  const text = newCommentText.value.trim()
  if (!text) return
  const id = Number(props.articleId)
  if (!id || Number.isNaN(id)) {
    error.value = "当前未绑定文章 ID，无法调用 /articles/{id}/comments"
    return
  }
  posting.value = true
  error.value = ""
  try {
    await createArticleComment(id, { content: text, parent_id: 0 })
    await loadComments()
  } catch (e) {
    error.value = e?.message || "发布评论失败"
  } finally {
    posting.value = false
  }
  newCommentText.value = ""
}

async function onPublishReply({ parentId, content }) {
  const text = String(content || "").trim()
  if (!text) return
  const id = Number(props.articleId)
  if (!id || Number.isNaN(id)) {
    error.value = "当前未绑定文章 ID，无法调用 /articles/{id}/comments"
    return
  }
  posting.value = true
  error.value = ""
  try {
    await createArticleComment(id, {
      content: text,
      parent_id: Number(parentId) || 0
    })
    await loadComments()
  } catch (e) {
    error.value = e?.message || "发布回复失败"
  } finally {
    posting.value = false
  }
}
</script>

<template>
  <section class="comment-panel">
    <header class="head">
      <h3 class="title">{{ props.title }}</h3>
      <div class="head-right">
        <span class="count">{{ commentCount }} 条</span>
        <button type="button" class="toggle-btn" @click="toggleComments">
          {{ collapsed ? "展开评论" : "收起评论" }}
        </button>
      </div>
    </header>

    <transition name="collapse">
      <div v-show="!collapsed" class="list-wrap">
        <div class="publish-box">
          <textarea
            v-model="newCommentText"
            class="publish-input"
            rows="3"
            placeholder="写下你的评论..."
            :disabled="posting"
          ></textarea>
          <div class="publish-actions">
            <button type="button" class="publish-btn" :disabled="posting" @click="publishComment">
              {{ posting ? "发布中..." : "发布评论" }}
            </button>
          </div>
        </div>

        <p v-if="!Number(props.articleId)" class="loading-tip">请先打开具体文章后再查看评论。</p>
        <p v-else-if="loading" class="loading-tip">评论加载中…</p>
        <p v-else-if="error" class="error-tip">{{ error }}</p>

        <ul v-if="commentCount && !loading" class="list">
          <CommentThread
            v-for="item in localComments"
            :key="item.id"
            :node="item"
            :level="0"
            @publish-reply="onPublishReply"
          />
        </ul>
        <p v-else class="empty">暂无评论，快来发表第一条评论吧。</p>
      </div>
    </transition>
  </section>
</template>

<style scoped>
.comment-panel {
  width: min(960px, 96vw);
  margin: 24px auto;
  border-radius: 14px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  background: rgba(15, 23, 42, 0.55);
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.35);
  overflow: hidden;
}

.head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.16);
}

.title {
  margin: 0;
  color: #f1f5f9;
  font-size: 16px;
}

.head-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.count {
  font-size: 12px;
  color: #94a3b8;
}

.toggle-btn {
  border: 1px solid rgba(148, 163, 184, 0.22);
  background: rgba(30, 41, 59, 0.7);
  color: #e2e8f0;
  border-radius: 8px;
  padding: 6px 10px;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
}

.toggle-btn:hover {
  background: rgba(51, 65, 85, 0.86);
}

.list-wrap {
  padding: 14px 16px;
}

.publish-box {
  margin-bottom: 12px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background: rgba(2, 6, 23, 0.4);
  border-radius: 10px;
  padding: 10px;
}

.publish-input {
  width: 100%;
  box-sizing: border-box;
  resize: vertical;
  min-height: 72px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 8px;
  background: rgba(15, 23, 42, 0.7);
  color: #e2e8f0;
  padding: 10px;
  outline: none;
}

.publish-input:focus {
  border-color: rgba(96, 165, 250, 0.45);
}

.publish-actions {
  margin-top: 8px;
  display: flex;
  justify-content: flex-end;
}

.publish-btn {
  border: 1px solid rgba(96, 165, 250, 0.35);
  background: rgba(37, 99, 235, 0.3);
  color: #dbeafe;
  border-radius: 8px;
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
}

.publish-btn:hover {
  background: rgba(37, 99, 235, 0.5);
}

.publish-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.loading-tip {
  margin: 0 0 8px;
  color: #94a3b8;
  font-size: 13px;
}

.error-tip {
  margin: 0 0 8px;
  color: #fca5a5;
  font-size: 13px;
}

.list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 12px;
}

.empty {
  margin: 0;
  color: #94a3b8;
  font-size: 14px;
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
