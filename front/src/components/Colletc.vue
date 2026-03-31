<script setup>
import { onMounted, ref } from "vue"
import Item from "./Item.vue"
import { getAccessToken } from "@/api/http.js"
import {
  fetchMyCollections,
  interactionRowToArticle,
  uncollectArticle
} from "@/api/blogApi.js"

const emit = defineEmits(["open-read"])

const articles = ref([])
const loading = ref(false)
const error = ref("")
const needLogin = ref(false)
const actionBusyId = ref(null)

async function load() {
  error.value = ""
  articles.value = []
  if (!getAccessToken()) {
    needLogin.value = true
    return
  }
  needLogin.value = false
  loading.value = true
  try {
    const rows = await fetchMyCollections(1, 100)
    articles.value = rows
      .map((r) => interactionRowToArticle(r, "收藏"))
      .filter(Boolean)
  } catch (e) {
    error.value = e?.message || "加载收藏失败"
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  load()
})

function toItem(a) {
  return {
    id: a.id,
    avatar: a.cover || "",
    name: a.title || "未命名",
    desc: a.summary || "我的收藏",
    tag: a.date || "收藏"
  }
}

function onSelectItem(it) {
  const id = it?.id
  if (id != null) emit("open-read", { id })
}

async function removeCollect(articleId) {
  if (!articleId) return
  actionBusyId.value = articleId
  error.value = ""
  try {
    // swagger: DELETE /articles/{id}/collections
    await uncollectArticle(articleId)
    // 先本地移除，再由刷新按钮可重新与后端对齐
    articles.value = articles.value.filter((a) => Number(a.id) !== Number(articleId))
  } catch (e) {
    error.value = e?.message || "取消收藏失败"
  } finally {
    actionBusyId.value = null
  }
}
</script>

<template>
  <section class="my-panel">
    <header class="head">
      <h2 class="title">我的收藏</h2>
      <button type="button" class="refresh" :disabled="loading" @click="load">
        {{ loading ? "加载中…" : "刷新" }}
      </button>
    </header>

    <p v-if="needLogin" class="hint warn">请先登录后查看收藏列表。</p>
    <p v-else-if="error" class="hint error">{{ error }}</p>
    <p v-else-if="loading" class="hint">加载中…</p>
    <p v-else-if="!articles.length" class="hint">暂无收藏。</p>

    <div v-else class="list">
      <div v-for="a in articles" :key="a.id" class="row">
        <Item :item="toItem(a)" @click="onSelectItem" />
        <div class="row-actions">
          <button
            type="button"
            class="open-btn"
            @click="onSelectItem({ id: a.id })"
          >
            阅读
          </button>
          <button
            type="button"
            class="remove-btn"
            :disabled="actionBusyId === a.id"
            @click="removeCollect(a.id)"
          >
            {{ actionBusyId === a.id ? "取消中..." : "取消收藏" }}
          </button>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.my-panel {
  width: min(980px, 96vw);
  margin: 24px auto;
  padding: 18px;
  border-radius: 16px;
  background: rgba(2, 6, 23, 0.45);
  border: 1px solid rgba(148, 163, 184, 0.14);
}

.head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.12);
}

.title {
  margin: 0;
  font-size: 17px;
  font-weight: 800;
  color: #f1f5f9;
}

.refresh {
  padding: 8px 14px;
  border-radius: 10px;
  border: 1px solid rgba(250, 204, 21, 0.35);
  background: rgba(113, 63, 18, 0.35);
  color: #fde68a;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}

.refresh:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.hint {
  margin: 12px 4px;
  color: #94a3b8;
  font-size: 14px;
}

.hint.error {
  color: #fca5a5;
}

.hint.warn {
  color: #fcd34d;
}

.list {
  display: grid;
  gap: 12px;
}

.row {
  display: grid;
  gap: 8px;
}

.row-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.open-btn,
.remove-btn {
  border-radius: 8px;
  padding: 6px 10px;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
}

.open-btn {
  border: 1px solid rgba(96, 165, 250, 0.35);
  background: rgba(30, 64, 175, 0.25);
  color: #bfdbfe;
}

.open-btn:hover {
  background: rgba(30, 64, 175, 0.4);
}

.remove-btn {
  border: 1px solid rgba(248, 113, 113, 0.4);
  background: rgba(127, 29, 29, 0.35);
  color: #fecaca;
}

.remove-btn:hover:enabled {
  background: rgba(127, 29, 29, 0.5);
}

.remove-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
</style>
