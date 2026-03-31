<script setup>
import { computed, ref } from "vue"
import Item from "./Item.vue"
import Read from "./Read.vue"

const props = defineProps({
  articles: {
    type: Array,
    default: () => []
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(["select"])

const selectedArticle = ref(null)

function escapeHtml(text) {
  return String(text || "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
}

/** 将列表项转为 Read 所需的 article（优先使用后端/父级传入的 HTML 正文） */
function articleToReadArticle(article) {
  let contentHtml
  if (article.contentHtml) {
    contentHtml = article.contentHtml
  } else if (article.content) {
    contentHtml = `<pre class="draft-raw-md">${escapeHtml(article.content)}</pre>`
  } else if (article.summary) {
    contentHtml = `<p>${escapeHtml(article.summary)}</p>`
  } else {
    contentHtml = `<p class="draft-empty">暂无正文</p>`
  }

  return {
    id: article.id,
    title: article.title || "未命名文章",
    subtitle: article.subtitle || "",
    author: article.author || "",
    category: article.category || "",
    publishedAt: article.date || article.publishedAt || "",
    readTime: article.readTime || "",
    cover: article.cover || "",
    contentHtml
  }
}

const readArticle = computed(() =>
  selectedArticle.value ? articleToReadArticle(selectedArticle.value) : null
)

function toItem(article) {
  return {
    avatar: article.cover || "",
    name: article.title || "未命名文章",
    desc: article.summary || "暂无摘要",
    tag: `${article.category || "未分类"} · ${article.date || "未知日期"}`
  }
}

function handleClick(article) {
  selectedArticle.value = article
  emit("select", article)
}

function closeRead() {
  selectedArticle.value = null
}
</script>

<template>
  <div class="article-list-page">
    <section v-show="!selectedArticle" class="panel">
      <header class="panel-head">
        <h3 class="panel-title">我的文章</h3>
      </header>

      <p v-if="props.loading" class="hint">加载中…</p>
      <p v-else-if="!props.articles.length" class="hint">
        暂无文章。请登录后由 GET /articles/author/{id} 拉取；若用户 ID 未就绪也会为空。
      </p>
      <div v-else class="list">
        <Item
          v-for="article in props.articles"
          :key="article.id"
          :item="toItem(article)"
          @click="handleClick(article)"
        />
      </div>
    </section>

    <div v-show="selectedArticle" class="article-read-wrap">
      <button type="button" class="back-btn" @click="closeRead">← 返回文章列表</button>
      <Read v-if="readArticle" :article="readArticle" />
    </div>
  </div>
</template>

<style scoped>
.article-list-page {
  width: 100%;
}

.panel {
  width: min(980px, 96vw);
  margin: 24px auto;
  padding: 18px;
  border-radius: 16px;
  background: rgba(2, 6, 23, 0.45);
  border: 1px solid rgba(148, 163, 184, 0.14);
}

.panel-head {
  padding: 6px 6px 12px;
}

.panel-title {
  margin: 0;
  font-size: 16px;
  font-weight: 800;
  color: #e5e7eb;
}

.hint {
  margin: 12px 6px;
  color: #94a3b8;
  font-size: 14px;
}

.list {
  display: grid;
  gap: 12px;
}

.article-read-wrap {
  width: 100%;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  margin: 0 0 16px;
  padding: 8px 14px;
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.28);
  background: rgba(30, 41, 59, 0.75);
  color: #e2e8f0;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}

.back-btn:hover {
  border-color: rgba(96, 165, 250, 0.4);
  background: rgba(51, 65, 85, 0.85);
}
</style>
