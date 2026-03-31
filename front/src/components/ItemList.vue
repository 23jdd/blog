<script setup>
const props = defineProps({
  articles: {
    type: Array,
    default: () => []
  },
  loading: {
    type: Boolean,
    default: false
  },
  error: {
    type: String,
    default: ""
  }
})

const emit = defineEmits(["change", "open-read"])

function openRead(article) {
  emit("open-read", article)
}
</script>

<template>
  <section class="article-panel">
    <header class="panel-head">
      <h3 class="panel-title">推荐</h3>
      <button class="ghost-btn" type="button">+ New Article</button>
    </header>

    <p v-if="loading" class="hint">加载中…</p>
    <p v-else-if="error" class="hint error">{{ error }}</p>
    <p v-else-if="!props.articles.length" class="hint">暂无推荐文章（请登录后重试或检查后端 /articles/hot、/interactions/feed）。</p>

    <div v-else class="list">
      <article
        v-for="article in props.articles"
        :key="article.id"
        class="article-card"
        role="button"
        tabindex="0"
        @click="openRead(article)"
        @keydown.enter.prevent="openRead(article)"
      >
        <div class="meta-top">
          <span class="badge">{{ article.category }}</span>
          <span class="date">{{ article.date }}</span>
        </div>

        <h4 class="title">{{ article.title }}</h4>
        <p class="summary">{{ article.summary }}</p>

        <div class="meta-bottom">
          <span v-if="article.author">作者：{{ article.author }}</span>
          <span v-if="article.readTime">阅读：{{ article.readTime }}</span>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.article-panel {
  width: min(980px, 96vw);
  margin: 24px auto;
  padding: 18px 18px 22px;
  border-radius: 16px;
  background: rgba(2, 6, 23, 0.45);
  border: 1px solid rgba(148, 163, 184, 0.14);
  box-shadow: 0 22px 60px rgba(2, 6, 23, 0.5);
}

.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 6px 6px 14px;
}

.panel-title {
  margin: 0;
  font-size: 16px;
  font-weight: 800;
  letter-spacing: 0.3px;
  color: #e5e7eb;
}

.hint {
  margin: 12px 6px;
  color: #94a3b8;
  font-size: 14px;
}

.hint.error {
  color: #fca5a5;
}

.ghost-btn {
  border-radius: 999px;
  padding: 8px 12px;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  color: #cbd5e1;
  background: rgba(30, 41, 59, 0.45);
  border: 1px solid rgba(148, 163, 184, 0.18);
  transition: background-color 0.16s ease, border-color 0.16s ease;
}

.ghost-btn:hover {
  background: rgba(30, 41, 59, 0.7);
  border-color: rgba(96, 165, 250, 0.28);
}

.list {
  display: grid;
  gap: 12px;
}

.article-card {
  border-radius: 12px;
  padding: 14px 14px 12px;
  background: rgba(15, 23, 42, 0.58);
  border: 1px solid rgba(148, 163, 184, 0.18);
  display: grid;
  gap: 10px;
  transition: transform 0.16s ease, border-color 0.16s ease;
  cursor: pointer;
}

.article-card:hover {
  transform: translateY(-2px);
  border-color: rgba(96, 165, 250, 0.3);
}

.meta-top,
.meta-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.badge {
  font-size: 12px;
  color: #93c5fd;
  background: rgba(37, 99, 235, 0.16);
  border: 1px solid rgba(96, 165, 250, 0.28);
  padding: 2px 8px;
  border-radius: 999px;
}

.date,
.meta-bottom {
  font-size: 12px;
  color: #9ca3af;
}

.title {
  margin: 0;
  font-size: 17px;
  color: #e5e7eb;
}

.summary {
  margin: 0;
  color: #cbd5e1;
  line-height: 1.6;
}

@media (max-width: 640px) {
  .meta-top,
  .meta-bottom {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
