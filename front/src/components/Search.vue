<script setup>
import { computed, ref, watch } from "vue"
import {
  searchArticlesRemote,
  articleModelToCard
} from "@/api/blogApi.js"
import Item from "./Item.vue"

const props = defineProps({
  pageSize: {
    type: Number,
    default: 8
  }
})

const emit = defineEmits(["select"])

const keyword = ref("")
const debouncedKeyword = ref("")
const page = ref(1)
const articles = ref([])
const loading = ref(false)
const error = ref("")
let debounceTimer = null

watch(keyword, (q) => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    debouncedKeyword.value = q.trim()
    page.value = 1
  }, 360)
})

watch([debouncedKeyword, page], async () => {
  const q = debouncedKeyword.value
  if (!q) {
    articles.value = []
    error.value = ""
    return
  }
  loading.value = true
  error.value = ""
  try {
    const raw = await searchArticlesRemote(q, page.value, props.pageSize)
    articles.value = raw
      .map((a) => articleModelToCard(a))
      .filter(Boolean)
  } catch (e) {
    error.value = e?.message || "搜索失败"
    articles.value = []
  } finally {
    loading.value = false
  }
})

const totalHint = computed(() => articles.value.length)

const totalPages = computed(() => {
  if (articles.value.length < props.pageSize) return page.value
  return page.value + 1
})

function goPrev() {
  page.value = Math.max(1, page.value - 1)
}

function goNext() {
  if (articles.value.length >= props.pageSize) page.value += 1
}

function selectArticle(article) {
  emit("select", article)
}

function toItem(article) {
  return {
    avatar: article.cover || "",
    name: article.title || "未命名文章",
    desc: article.summary || "暂无摘要",
    tag: article.category || "未分类"
  }
}
</script>

<template>
  <section class="search-panel">
    <header class="head">
      <h3 class="title">搜索文章</h3>
      <p class="count">
        <span v-if="loading">搜索中…</span>
        <span v-else>本页 {{ totalHint }} 条（GET /articles/search）</span>
      </p>
    </header>

    <div class="search-row">
      <input
        v-model="keyword"
        type="text"
        class="search-input"
        placeholder="输入关键词（服务端搜索）..."
      />
    </div>

    <p v-if="error" class="empty error">{{ error }}</p>

    <div v-if="articles.length" class="result-list">
      <div v-for="article in articles" :key="article.id" class="result-row">
        <Item :item="toItem(article)" @click="selectArticle(article)" />
        <p class="meta">{{ article.date }}</p>
      </div>
    </div>
    <p v-else-if="debouncedKeyword && !loading" class="empty">没有匹配的文章，请换个关键词试试。</p>
    <p v-else-if="!debouncedKeyword" class="empty">输入关键词开始搜索。</p>

    <footer v-if="debouncedKeyword" class="pager">
      <button type="button" class="page-btn" :disabled="page === 1 || loading" @click="goPrev">
        上一页
      </button>
      <span class="page-text">第 {{ page }} 页</span>
      <button
        type="button"
        class="page-btn"
        :disabled="loading || articles.length < pageSize"
        @click="goNext"
      >
        下一页
      </button>
    </footer>
  </section>
</template>

<style scoped>
.search-panel {
  width: min(980px, 96vw);
  margin: 24px auto;
  padding: 18px;
  border-radius: 14px;
  background: rgba(15, 23, 42, 0.55);
  border: 1px solid rgba(148, 163, 184, 0.18);
}

.head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.title {
  margin: 0;
  font-size: 17px;
  color: #f1f5f9;
}

.count {
  margin: 0;
  color: #94a3b8;
  font-size: 12px;
}

.search-row {
  margin-bottom: 12px;
}

.search-input {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid rgba(148, 163, 184, 0.24);
  border-radius: 10px;
  background: rgba(2, 6, 23, 0.5);
  color: #e2e8f0;
  padding: 10px 12px;
  outline: none;
}

.search-input:focus {
  border-color: rgba(96, 165, 250, 0.45);
}

.result-list {
  display: grid;
  gap: 10px;
}

.result-row {
  display: grid;
  gap: 6px;
}

.meta {
  margin: 0;
  color: #94a3b8;
  font-size: 12px;
}

.empty {
  margin: 12px 0;
  color: #94a3b8;
  font-size: 14px;
}

.empty.error {
  color: #fca5a5;
}

.pager {
  margin-top: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.page-btn {
  border: 1px solid rgba(148, 163, 184, 0.24);
  background: rgba(30, 41, 59, 0.7);
  color: #e2e8f0;
  border-radius: 8px;
  padding: 6px 10px;
  cursor: pointer;
}

.page-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.page-text {
  color: #cbd5e1;
  font-size: 13px;
}
</style>
