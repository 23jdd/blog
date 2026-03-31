<script setup>
import { computed, ref, watch } from "vue"
import Recommand from "./Recommand.vue"
import { getAccessToken } from "@/api/http.js"
import {
  fetchArticleById,
  fetchTextUrl,
  renderMarkdown,
  resolveMediaUrl,
  formatDraftTimeLabel,
  fetchArticleStats,
  fetchMyCollections,
  likeArticle,
  unlikeArticle,
  collectArticle,
  uncollectArticle,
  followUser,
  unfollowUser
} from "@/api/blogApi.js"

const DEMO_ARTICLE = {
  title: "Vue 3 组合式 API 与可维护前端架构",
  subtitle: "从组件边界到状态管理的一次实践记录",
  author: "Long",
  category: "前端",
  publishedAt: "2026-03-26",
  readTime: "约 8 分钟",
  cover: "",
  contentHtml: `
<p>这篇文章用于演示 <strong>Read</strong> 阅读页的布局与排版。你可以通过父组件传入 <code>article</code> 来替换标题与正文。</p>
<h2>为什么要拆分组件</h2>
<p>清晰的边界能让页面更易测试、复用与迭代。组合式 API 让相关逻辑可以收敛在同一块代码里，而不是散落在多个选项字段中。</p>
<ul>
  <li>单一职责：展示与数据获取分层</li>
  <li>可测试：纯展示组件更容易快照与单元测试</li>
  <li>可扩展：后续接入路由或接口时改动面更小</li>
</ul>
<pre><code>// 示例：父组件传入后端渲染好的 HTML
&lt;Read :article="{ title: '...', contentHtml: html }" /&gt;</code></pre>
<p>正文支持通过 <code>contentHtml</code> 传入服务端渲染的 Markdown HTML，与草稿编辑器产出的内容衔接。</p>
`
}

const props = defineProps({
  /** 设置后通过 GET /articles/{id} 拉取详情并渲染（swagger: model.Article） */
  articleId: {
    type: [Number, String],
    default: null
  },
  /** 直接展示（如 ArticleList 内嵌）；与 articleId 同时存在时 articleId 优先 */
  article: {
    type: Object,
    default: null
  }
})

const loading = ref(false)
const loadError = ref("")
const remoteArticle = ref(null)

watch(
  () => props.articleId,
  async (id) => {
    const n = id != null && id !== "" ? Number(id) : NaN
    if (Number.isNaN(n) || n <= 0) {
      remoteArticle.value = null
      loadError.value = ""
      return
    }
    loading.value = true
    loadError.value = ""
    try {
      const a = await fetchArticleById(n)
      const contentField = a.content != null ? String(a.content) : ""
      let contentHtml = ""
      if (
        contentField &&
        (contentField.includes("/") || /\.(md|markdown|txt)$/i.test(contentField))
      ) {
        try {
          const md = await fetchTextUrl(contentField)
          contentHtml = await renderMarkdown(md)
        } catch {
          contentHtml = `<p class="draft-empty">无法加载正文资源：${escapeAttr(
            contentField
          )}</p>`
        }
      } else if (contentField.trim()) {
        contentHtml = await renderMarkdown(contentField)
      } else {
        contentHtml = '<p class="draft-empty">暂无正文</p>'
      }
      const tags = a.tags != null ? String(a.tags) : ""
      remoteArticle.value = {
        title: a.title || "未命名文章",
        subtitle: tags,
        author: a.author_name || a.author || a.username || "",
        authorId: a.author_id ?? null,
        category: tags.split(",")[0]?.trim() || "",
        publishedAt: a.create_time || a.update_time || "",
        readTime: "",
        cover: resolveMediaUrl(a.cover_url || ""),
        contentHtml
      }
    } catch (e) {
      loadError.value = e?.message || "加载文章失败"
      remoteArticle.value = null
    } finally {
      loading.value = false
    }
  },
  { immediate: true }
)

function escapeAttr(s) {
  return String(s)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/"/g, "&quot;")
}

const displayArticle = computed(() => {
  if (remoteArticle.value) return remoteArticle.value
  if (props.article && typeof props.article === "object" && props.article.title != null) {
    return props.article
  }
  return DEMO_ARTICLE
})

function formatDate(iso) {
  if (!iso) return ""
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return formatDraftTimeLabel(iso) || String(iso)
  return d.toLocaleDateString("zh-CN", {
    year: "numeric",
    month: "long",
    day: "numeric"
  })
}

const metaParts = computed(() => {
  const a = displayArticle.value
  const parts = []
  if (a.author) parts.push(`作者：${a.author}`)
  if (a.publishedAt) parts.push(formatDate(a.publishedAt))
  if (a.readTime) parts.push(a.readTime)
  return parts
})

/** 可用于点赞/收藏的文章 id（路由阅读用 articleId，内嵌用 article.id） */
const interactionArticleId = computed(() => {
  const p = props.articleId
  const n1 = p != null && p !== "" ? Number(p) : NaN
  if (!Number.isNaN(n1) && n1 > 0) return n1
  const aid = props.article?.id
  const n2 = aid != null ? Number(aid) : NaN
  if (!Number.isNaN(n2) && n2 > 0) return n2
  return null
})

const hasToken = computed(() => !!getAccessToken())
const liked = ref(false)
const collected = ref(false)
const likeCount = ref(null)
const interactionBusy = ref(false)
const isFollowingAuthor = ref(false)
const followBusy = ref(false)

const followTargetId = computed(() => {
  const id = displayArticle.value?.authorId
  const n = Number(id)
  if (Number.isNaN(n) || n <= 0) return null
  return n
})

watch(
  followTargetId,
  () => {
    // 后端目前未提供“查询我是否已关注某用户”的接口，默认 false，点击后切换
    isFollowingAuthor.value = false
  },
  { immediate: true }
)

function normalizeStatsPayload(raw) {
  if (raw == null || typeof raw !== "object") return {}
  return raw
}

function pickLikedFromResponse(raw) {
  const o = raw && typeof raw === "object" ? raw : {}
  if (typeof o.liked === "boolean") return o.liked
  const d = o.data
  if (d && typeof d === "object" && typeof d.liked === "boolean") return d.liked
  return null
}

watch(
  interactionArticleId,
  async (id) => {
    liked.value = false
    collected.value = false
    likeCount.value = null
    if (id == null || !hasToken.value) return
    try {
      const st = await fetchArticleStats(id)
      const s = normalizeStatsPayload(st)
      const lc = s.like_count ?? s.likeCount ?? s.likes
      if (typeof lc === "number") likeCount.value = lc
      if (typeof s.liked === "boolean") liked.value = s.liked
      if (typeof s.collected === "boolean") collected.value = s.collected
    } catch {
      /* 统计接口可选 */
    }
    try {
      const cols = await fetchMyCollections(1, 200)
      const hit = cols.some(
        (c) => Number(c.articleID ?? c.article_id) === Number(id)
      )
      collected.value = hit
    } catch {
      /* 未登录或接口失败时保持 false */
    }
  },
  { immediate: true }
)

async function refreshLikeCount() {
  const id = interactionArticleId.value
  if (id == null || !hasToken.value) return
  try {
    const st = await fetchArticleStats(id)
    const s = normalizeStatsPayload(st)
    const lc = s.like_count ?? s.likeCount ?? s.likes
    if (typeof lc === "number") likeCount.value = lc
  } catch {
    /* ignore */
  }
}

async function toggleLike() {
  const id = interactionArticleId.value
  if (id == null) return
  if (!hasToken.value) {
    window.alert("请先登录后再点赞。")
    return
  }
  interactionBusy.value = true
  try {
    if (liked.value) {
      const r = await unlikeArticle(id)
      const next = pickLikedFromResponse(r)
      liked.value = typeof next === "boolean" ? next : false
    } else {
      const r = await likeArticle(id)
      const next = pickLikedFromResponse(r)
      liked.value = typeof next === "boolean" ? next : true
    }
    await refreshLikeCount()
  } catch (e) {
    window.alert(e?.message || "点赞操作失败")
  } finally {
    interactionBusy.value = false
  }
}

async function toggleCollect() {
  const id = interactionArticleId.value
  if (id == null) return
  if (!hasToken.value) {
    window.alert("请先登录后再收藏。")
    return
  }
  interactionBusy.value = true
  try {
    if (collected.value) {
      await uncollectArticle(id)
      collected.value = false
    } else {
      await collectArticle(id)
      collected.value = true
    }
  } catch (e) {
    window.alert(e?.message || "收藏操作失败")
  } finally {
    interactionBusy.value = false
  }
}

async function toggleFollowAuthor() {
  const targetID = followTargetId.value
  if (targetID == null) return
  if (!hasToken.value) {
    window.alert("请先登录后再关注作者。")
    return
  }
  followBusy.value = true
  try {
    if (isFollowingAuthor.value) {
      await unfollowUser(targetID)
      isFollowingAuthor.value = false
    } else {
      await followUser(targetID)
      isFollowingAuthor.value = true
    }
  } catch (e) {
    window.alert(e?.message || "关注操作失败")
  } finally {
    followBusy.value = false
  }
}
</script>

<template>
  <article class="read-page">
    <p v-if="loading" class="read-status">加载文章中…</p>
    <p v-else-if="loadError" class="read-status error">{{ loadError }}</p>

    <template v-else>
      <header class="read-header">
        <div v-if="displayArticle.category" class="category">{{ displayArticle.category }}</div>
        <h1 class="read-title">{{ displayArticle.title }}</h1>
        <p v-if="displayArticle.subtitle" class="read-subtitle">{{ displayArticle.subtitle }}</p>

        <p v-if="metaParts.length" class="meta">{{ metaParts.join(" · ") }}</p>

        <div v-if="interactionArticleId" class="read-actions">
          <button
            v-if="followTargetId"
            type="button"
            class="action-btn action-follow"
            :class="{ active: isFollowingAuthor }"
            :disabled="followBusy"
            :aria-pressed="isFollowingAuthor"
            @click="toggleFollowAuthor"
          >
            <span class="action-icon" aria-hidden="true">{{ isFollowingAuthor ? "✓" : "+" }}</span>
            {{ isFollowingAuthor ? "已关注作者" : "关注作者" }}
          </button>
          <button
            type="button"
            class="action-btn action-like"
            :class="{ active: liked }"
            :disabled="interactionBusy"
            :aria-pressed="liked"
            @click="toggleLike"
          >
            <span class="action-icon" aria-hidden="true">{{ liked ? "♥" : "♡" }}</span>
            {{ liked ? "已赞" : "点赞" }}
            <span v-if="likeCount != null" class="action-count">{{ likeCount }}</span>
          </button>
          <button
            type="button"
            class="action-btn action-collect"
            :class="{ active: collected }"
            :disabled="interactionBusy"
            :aria-pressed="collected"
            @click="toggleCollect"
          >
            <span class="action-icon" aria-hidden="true">{{ collected ? "★" : "☆" }}</span>
            {{ collected ? "已收藏" : "收藏" }}
          </button>
          <span v-if="!hasToken" class="action-hint">登录后可点赞与收藏</span>
        </div>
      </header>

      <figure v-if="displayArticle.cover" class="cover-wrap">
        <img :src="displayArticle.cover" alt="" class="cover-img" />
      </figure>

      <div class="read-body prose" v-html="displayArticle.contentHtml"></div>
    </template>
  </article>
  <section class="recommand-section">
    <Recommand :article-id="interactionArticleId" />
  </section>
</template>

<style scoped>
.read-status {
  margin: 0 0 16px;
  color: #94a3b8;
  font-size: 15px;
}

.read-status.error {
  color: #fca5a5;
}

.read-page {
  box-sizing: border-box;
  width: 100%;
  min-height: 100vh;
  min-height: 100dvh;
  margin: 0;
  padding: clamp(20px, 4vw, 48px) clamp(16px, 5vw, 64px) 56px;
  border-radius: 0;
  background: linear-gradient(165deg, #1e293b 0%, #0f172a 45%, #0c1220 100%);
  border: none;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.06);
  display: flex;
  flex-direction: column;
}

.read-header {
  margin-bottom: 28px;
  padding-bottom: 22px;
  border-bottom: 1px solid rgba(226, 232, 240, 0.12);
}

.category {
  display: inline-block;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #bfdbfe;
  background: rgba(59, 130, 246, 0.22);
  border: 1px solid rgba(147, 197, 253, 0.45);
  padding: 5px 12px;
  border-radius: 999px;
  margin-bottom: 16px;
}

.read-title {
  margin: 0 0 12px;
  font-size: clamp(1.55rem, 4vw, 2.1rem);
  font-weight: 800;
  line-height: 1.22;
  color: #ffffff;
  letter-spacing: 0.02em;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.35);
}

.read-subtitle {
  margin: 0 0 18px;
  font-size: 1.08rem;
  line-height: 1.55;
  color: #cbd5e1;
  font-weight: 500;
}

.meta {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #e2e8f0;
  line-height: 1.6;
}

.read-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  margin-top: 18px;
  padding-top: 18px;
  border-top: 1px solid rgba(226, 232, 240, 0.1);
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  border: 1px solid rgba(148, 163, 184, 0.35);
  background: rgba(15, 23, 42, 0.65);
  color: #cbd5e1;
  transition:
    background 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.action-btn:hover:not(:disabled) {
  border-color: rgba(96, 165, 250, 0.45);
  background: rgba(30, 41, 59, 0.85);
  color: #f1f5f9;
}

.action-like:hover:not(:disabled):not(.active) {
  border-color: rgba(248, 113, 113, 0.35);
  color: #fca5a5;
}

.action-collect:hover:not(:disabled):not(.active) {
  border-color: rgba(250, 204, 21, 0.35);
  color: #fde047;
}

.action-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.action-follow.active {
  border-color: rgba(74, 222, 128, 0.5);
  background: rgba(21, 128, 61, 0.28);
  color: #bbf7d0;
}

/* 已点赞：明确红色系 */
.action-like.active {
  border-color: #f87171;
  background: linear-gradient(145deg, rgba(185, 28, 28, 0.55), rgba(127, 29, 29, 0.45));
  color: #fecaca;
  box-shadow:
    0 0 0 1px rgba(248, 113, 113, 0.25),
    0 4px 18px rgba(220, 38, 38, 0.28);
}

.action-like.active:hover:not(:disabled) {
  border-color: #fca5a5;
  background: linear-gradient(145deg, rgba(220, 38, 38, 0.6), rgba(153, 27, 27, 0.5));
  color: #fff1f2;
}

.action-like.active .action-icon {
  color: #f87171;
  text-shadow: 0 0 12px rgba(248, 113, 113, 0.75);
}

.action-like.active .action-count {
  background: rgba(69, 10, 10, 0.65);
  color: #fecaca;
  border: 1px solid rgba(248, 113, 113, 0.35);
}

/* 已收藏：明确黄色 / 琥珀色 */
.action-collect.active {
  border-color: #fbbf24;
  background: linear-gradient(145deg, rgba(146, 64, 14, 0.55), rgba(113, 63, 18, 0.48));
  color: #fef9c3;
  box-shadow:
    0 0 0 1px rgba(250, 204, 21, 0.22),
    0 4px 18px rgba(234, 179, 8, 0.22);
}

.action-collect.active:hover:not(:disabled) {
  border-color: #fcd34d;
  background: linear-gradient(145deg, rgba(180, 83, 9, 0.58), rgba(120, 53, 15, 0.52));
  color: #fffbeb;
}

.action-collect.active .action-icon {
  color: #facc15;
  text-shadow: 0 0 12px rgba(250, 204, 21, 0.65);
}

.action-icon {
  font-size: 15px;
  line-height: 1;
  opacity: 0.95;
}

.action-count {
  margin-left: 2px;
  padding: 2px 7px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 800;
  background: rgba(0, 0, 0, 0.35);
  color: #e2e8f0;
}

.action-hint {
  font-size: 12px;
  color: #64748b;
  width: 100%;
}

@media (min-width: 520px) {
  .action-hint {
    width: auto;
    margin-left: auto;
  }
}

.cover-wrap {
  margin: 0 0 28px;
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid rgba(226, 232, 240, 0.18);
  background: rgba(15, 23, 42, 0.75);
  box-shadow: 0 12px 28px rgba(0, 0, 0, 0.35);
}

.cover-img {
  display: block;
  width: 100%;
  max-height: 360px;
  object-fit: cover;
}

.read-body {
  flex: 1 1 auto;
  width: 100%;
  max-width: 100%;
  font-size: 17px;
  line-height: 1.88;
  color: #f1f5f9;
  font-weight: 400;
}

.read-body :deep(strong) {
  color: #ffffff;
  font-weight: 700;
}

.read-body :deep(h2) {
  margin: 1.6em 0 0.65em;
  font-size: 1.4rem;
  font-weight: 800;
  color: #ffffff;
  letter-spacing: 0.01em;
}

.read-body :deep(h3) {
  margin: 1.35em 0 0.5em;
  font-size: 1.15rem;
  font-weight: 700;
  color: #f8fafc;
}

.read-body :deep(p) {
  margin: 0.85em 0;
}

.read-body :deep(ul),
.read-body :deep(ol) {
  margin: 0.85em 0;
  padding-left: 1.35em;
}

.read-body :deep(li) {
  margin: 0.35em 0;
}

.read-body :deep(a) {
  color: #7dd3fc;
  font-weight: 600;
  text-decoration: underline;
  text-decoration-color: rgba(125, 211, 252, 0.65);
  text-underline-offset: 3px;
}

.read-body :deep(code) {
  font-size: 0.9em;
  padding: 3px 8px;
  border-radius: 6px;
  background: rgba(51, 65, 85, 0.85);
  border: 1px solid rgba(148, 163, 184, 0.28);
  color: #f8fafc;
}

.read-body :deep(pre) {
  margin: 1.1em 0;
  padding: 16px 18px;
  border-radius: 10px;
  overflow: auto;
  background: #020617;
  border: 1px solid rgba(148, 163, 184, 0.28);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.04);
}

/* 草稿箱预览：整段 Markdown 源码 */
.read-body :deep(pre.draft-raw-md) {
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 13px;
  line-height: 1.55;
}

.read-body :deep(p.draft-empty) {
  color: #94a3b8;
  font-style: italic;
}

.read-body :deep(pre code) {
  padding: 0;
  background: transparent;
  border: none;
  font-size: 13px;
  line-height: 1.6;
  color: #e2e8f0;
}

.read-body :deep(blockquote) {
  margin: 1em 0;
  padding: 12px 18px;
  border-left: 4px solid #60a5fa;
  background: rgba(30, 41, 59, 0.65);
  color: #e2e8f0;
  border-radius: 0 10px 10px 0;
}

.read-body :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 8px;
  margin: 1em 0;
}
</style>
