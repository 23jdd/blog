<script setup>
import { computed, ref, watch } from "vue"
import { API_BASE } from "@/api/config.js"

const props = defineProps({
  /** 从草稿箱「编辑」打开时传入；含 _ts 用于每次点击都触发加载 */
  openFromDraftBox: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(["save-to-draft-list"])

/** 正在编辑的草稿箱条目 id，存入草稿箱时回传以便 Main 更新同一条 */
const editingDraftId = ref(null)

const articleTitle = ref("我的新文章")
/** 存入草稿箱时使用的标签（与草稿箱筛选一致） */
const draftTag = ref("前端")

const draftText = ref(`在这里输入正文内容，右侧会实时预览。

## 小标题

- 列表项 1
- 列表项 2

**加粗**、*斜体*、\`行内代码\`
`)

function escapeHtml(text) {
  return String(text || "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
}

function renderInline(text) {
  return text
    .replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>")
    .replace(/\*(.+?)\*/g, "<em>$1</em>")
    .replace(/`([^`]+?)`/g, "<code>$1</code>")
    .replace(
      /\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)/g,
      '<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>'
    )
}

const localPreviewHtml = computed(() => {
    
  const titleEscaped = escapeHtml(articleTitle.value.trim() || "未命名文章")
  const titleBlock = `<h1 class="article-title">${titleEscaped}</h1>`

  const lines = escapeHtml(draftText.value).split("\n")
  const html = []
  let inList = false

  for (const rawLine of lines) {
    const line = rawLine.trim()

    if (!line) {
      if (inList) {
        html.push("</ul>")
        inList = false
      }
      html.push("<p><br/></p>")
      continue
    }

    if (line.startsWith("- ")) {
      if (!inList) {
        html.push("<ul>")
        inList = true
      }
      html.push(`<li>${renderInline(line.slice(2))}</li>`)
      continue
    }

    if (inList) {
      html.push("</ul>")
      inList = false
    }

    if (line.startsWith("### ")) {
      html.push(`<h3>${renderInline(line.slice(4))}</h3>`)
    } else if (line.startsWith("## ")) {
      html.push(`<h2>${renderInline(line.slice(3))}</h2>`)
    } else if (line.startsWith("# ")) {
      html.push(`<h1>${renderInline(line.slice(2))}</h1>`)
    } else {
      html.push(`<p>${renderInline(line)}</p>`)
    }
  }

  if (inList) html.push("</ul>")
  return titleBlock + html.join("")
})

const serverPreviewHtml = ref("")
const isPreviewLoading = ref(false)
const previewError = ref("")

/** 从草稿箱打开时：挂载当下就要写入表单（须 immediate，否则首屏仍是默认模板且 editingDraftId 为空，保存会当成新草稿） */
function applyOpenDraftBox() {
  const d = props.openFromDraftBox
  if (!d || d._ts == null) return
  editingDraftId.value = d.id != null ? d.id : null
  articleTitle.value = d.title || "未命名草稿"
  draftTag.value = d.tag || "未分类"
  const body =
    d.content != null && String(d.content).trim() !== ""
      ? d.content
      : d.excerpt || ""
  draftText.value = body
  serverPreviewHtml.value = ""
  previewError.value = ""
}

watch(
  () => props.openFromDraftBox?._ts,
  applyOpenDraftBox,
  { immediate: true }
)

const previewHtml = computed(() => serverPreviewHtml.value || localPreviewHtml.value)

/** 解析 POST /markdown 返回（swagger SuccessResponse 常为 { data: { html } }） */
function pickMarkdownHtmlPayload(json) {
  if (json == null || typeof json !== "object") return ""
  const inner = json.data !== undefined ? json.data : json
  if (typeof inner === "string") return inner
  return inner?.html ?? inner?.HTML ?? json.html ?? json.HTML ?? ""
}

function insertTemplate() {
  draftText.value += `

## 新章节
这里输入你的想法...
`
}

function clearAll() {
  editingDraftId.value = null
  articleTitle.value = ""
  draftText.value = ""
  serverPreviewHtml.value = ""
  previewError.value = ""
}
async function requestServerPreview() {
  previewError.value = ""
  isPreviewLoading.value = true
  try {
    const content = `# ${articleTitle.value || "未命名文章"}\n\n${draftText.value}`
    const r = await fetch(`${API_BASE}/markdown`, {
      method: "POST",
      body: content,
      headers: {
        "Content-Type": "text/plain; charset=utf-8"
      }
    })
    const data = await r.json().catch(() => ({}))
    if (!r.ok) {
      throw new Error(data?.message || `预览失败（HTTP ${r.status}）`)
    }
    const html = pickMarkdownHtmlPayload(data)
    if (!html || !String(html).trim()) {
      throw new Error("后端未返回 HTML 正文（请检查响应中 data.html 或 html 字段）")
    }
    serverPreviewHtml.value = String(html)
  } catch (e) {
    previewError.value = e?.message || "请求后端预览失败"
  } finally {
    isPreviewLoading.value = false
  }
}

function saveDraft() {
  const payload = {
    title: articleTitle.value,
    content: draftText.value,
    tag: draftTag.value,
    savedAt: new Date().toISOString()
  }
  window.localStorage.setItem("draft-cache", JSON.stringify(payload))
}

function saveToDraftList() {
  const title = articleTitle.value.trim()
  const content = draftText.value.trim()
  if (!title && !content) {
    window.alert("请至少填写标题或正文后再存入草稿箱。")
    return
  }
  emit("save-to-draft-list", {
    id: editingDraftId.value,
    title: title || "未命名草稿",
    content: draftText.value,
    tag: draftTag.value || "未分类"
  })
  saveDraft()
}
</script>

<template>
  <section class="draft-wrap">
    <header class="head">
      <h3 class="title">Draft Editor</h3>
      <div class="actions">
     
        <button type="button" class="btn accent" @click="saveToDraftList">存入草稿箱</button>
        <button type="button" class="btn secondary" @click="insertTemplate">插入模板</button>
        <button type="button" class="btn secondary" @click="requestServerPreview">
          {{ isPreviewLoading ? "预览中..." : "后端预览" }}
        </button>
        <button type="button" class="btn danger" @click="clearAll">清空</button>

      </div>
    </header>

    <div class="title-row">
      <label class="title-label" for="draft-article-title">文章标题</label>
      <input
        id="draft-article-title"
        v-model="articleTitle"
        type="text"
        class="title-input"
        placeholder="输入文章标题..."
        maxlength="120"
      />
    </div>

    <div class="panes">
      <article class="pane">
        <h4 class="pane-title">编辑器</h4>
        <textarea
          v-model="draftText"
          class="editor"
          placeholder="在这里输入 Markdown 内容..."
        ></textarea>
      </article>

      <article class="pane pane-preview">
        <div class="pane-head">
          <h4 class="pane-title">
            <span class="pane-title-icon" aria-hidden="true" />
            预览
          </h4>
          <span class="preview-badge" :class="{ live: !serverPreviewHtml, server: !!serverPreviewHtml }">
            {{ serverPreviewHtml ? "后端渲染" : "实时预览" }}
          </span>
        </div>
        <div class="preview-body">
          <p v-if="previewError" class="error-tip">{{ previewError }}</p>
          <div class="preview-shell" :class="{ 'is-loading': isPreviewLoading }">
            <div v-if="isPreviewLoading" class="preview-loading" aria-live="polite">
              <span class="preview-loading-dot" />
              <span class="preview-loading-dot" />
              <span class="preview-loading-dot" />
              <span class="preview-loading-text">正在请求后端预览…</span>
            </div>
            <div
              class="preview markdown-body"
              :class="{ 'preview--from-server': !!serverPreviewHtml }"
              v-html="previewHtml"
            ></div>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.draft-wrap {
  width: min(1100px, 96vw);
  margin: 22px auto;
  padding: 16px;
  border-radius: 14px;
  background: rgba(15, 23, 42, 0.5);
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
  color: #e5e7eb;
  font-size: 16px;
}

.actions {
  display: flex;
  gap: 8px;
}

.title-row {
  display: grid;
  gap: 8px;
  margin-bottom: 12px;
}

.title-label {
  font-size: 13px;
  font-weight: 600;
  color: #94a3b8;
}

.title-input {
  width: 100%;
  box-sizing: border-box;
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.22);
  padding: 11px 14px;
  font-size: 15px;
  font-weight: 600;
  color: #f8fafc;
  background: rgba(15, 23, 42, 0.65);
  outline: none;
}

.title-input::placeholder {
  color: #64748b;
  font-weight: 500;
}

.title-input:focus {
  border-color: rgba(96, 165, 250, 0.55);
}

.panes {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.pane {
  background: rgba(2, 6, 23, 0.45);
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 12px;
  padding: 12px;
  min-height: 520px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.pane > .pane-title {
  flex-shrink: 0;
}

.pane > .editor {
  flex: 1 1 0;
  min-height: 280px;
}

.pane-preview {
  background: linear-gradient(
    165deg,
    #020617 0%,
    #0a0f18 45%,
    #020617 100%
  );
  border-color: rgba(51, 65, 85, 0.55);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.02),
    0 16px 48px rgba(0, 0, 0, 0.55);
}

.pane-head {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  flex-wrap: wrap;
}

.pane-title {
  margin: 0;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 700;
  color: #e2e8f0;
  letter-spacing: 0.02em;
}

.pane-preview .pane-title {
  color: #94a3b8;
}

.pane-title-icon {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.pane-preview .pane-title-icon {
  background: linear-gradient(135deg, #115e59, #164e63);
  box-shadow: 0 0 8px rgba(17, 94, 89, 0.22);
}

.preview-badge {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  padding: 4px 10px;
  border-radius: 999px;
  border: 1px solid rgba(51, 65, 85, 0.85);
  color: #64748b;
  background: rgba(2, 6, 23, 0.85);
}

.preview-badge.live {
  color: #5eead4;
  border-color: rgba(13, 148, 136, 0.35);
  background: rgba(2, 44, 42, 0.65);
}

.preview-badge.server {
  color: #a78bfa;
  border-color: rgba(91, 33, 182, 0.4);
  background: rgba(30, 27, 75, 0.55);
}

.preview-body {
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
}

.editor {
  width: 100%;
  height: 100%;
  resize: none;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 10px;
  padding: 12px;
  background: rgba(15, 23, 42, 0.75);
  color: #f3f4f6;
  line-height: 1.65;
  font-size: 14px;
  outline: none;
}

.editor:focus {
  border-color: rgba(96, 165, 250, 0.65);
}

.preview-shell {
  position: relative;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  border-radius: 12px;
  border: 1px solid rgba(30, 41, 59, 0.95);
  background: #010409;
  box-shadow:
    inset 0 0 0 1px rgba(0, 0, 0, 0.5),
    inset 0 1px 0 rgba(255, 255, 255, 0.03);
  overflow: hidden;
}

.preview-shell.is-loading .preview {
  opacity: 0.35;
  filter: blur(0.3px);
  pointer-events: none;
  transition: opacity 0.2s ease;
}

.preview-loading {
  position: absolute;
  inset: 0;
  z-index: 2;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  flex-wrap: wrap;
  background: rgba(0, 0, 0, 0.82);
  backdrop-filter: blur(6px);
}

.preview-loading-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #0f766e;
  animation: preview-dot 1.1s ease-in-out infinite;
  box-shadow: 0 0 8px rgba(15, 118, 110, 0.35);
}

.preview-loading-dot:nth-child(2) {
  animation-delay: 0.15s;
  background: #0369a1;
  box-shadow: 0 0 8px rgba(3, 105, 161, 0.35);
}

.preview-loading-dot:nth-child(3) {
  animation-delay: 0.3s;
  background: #5b21b6;
  box-shadow: 0 0 8px rgba(91, 33, 182, 0.35);
}

.preview-loading-text {
  width: 100%;
  text-align: center;
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
  padding: 0 12px;
}

@keyframes preview-dot {
  0%,
  80%,
  100% {
    transform: scale(0.65);
    opacity: 0.5;
  }
  40% {
    transform: scale(1.1);
    opacity: 1;
  }
}

.preview {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 18px 20px 22px;
  background: #020617;
  color: #94a3b8;
  line-height: 1.78;
  font-size: 15px;
  scrollbar-width: thin;
  scrollbar-color: rgba(51, 65, 85, 0.9) #020617;
}

.preview::-webkit-scrollbar {
  width: 8px;
}

.preview::-webkit-scrollbar-track {
  background: #020617;
  border-radius: 8px;
}

.preview::-webkit-scrollbar-thumb {
  background: linear-gradient(180deg, rgba(51, 65, 85, 0.95), rgba(30, 41, 59, 0.9));
  border-radius: 8px;
  border: 2px solid #020617;
}

.preview :deep(.article-title) {
  margin: 0 0 0.85em;
  padding-bottom: 0.7em;
  border-bottom: 1px solid rgba(51, 65, 85, 0.85);
  font-size: 1.55rem;
  font-weight: 800;
  letter-spacing: 0.02em;
  color: #e2e8f0;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.65);
}

.preview :deep(h1:not(.article-title)),
.preview :deep(h2),
.preview :deep(h3) {
  margin: 1em 0 0.5em;
  font-weight: 700;
  color: #cbd5e1;
  letter-spacing: 0.01em;
}

.preview :deep(h2) {
  font-size: 1.22rem;
  padding-bottom: 0.35em;
  border-bottom: 1px solid rgba(30, 41, 59, 0.95);
}

.preview :deep(h3) {
  font-size: 1.05rem;
  color: #94a3b8;
}

.preview :deep(p) {
  margin: 0.65em 0;
  color: #8696ac;
}

.preview :deep(ul),
.preview :deep(ol) {
  padding-left: 1.35em;
  margin: 0.6em 0;
}

.preview :deep(li) {
  margin: 0.35em 0;
  color: #8696ac;
}

.preview :deep(strong) {
  color: #e2e8f0;
  font-weight: 700;
}

.preview :deep(blockquote) {
  margin: 1em 0;
  padding: 12px 16px;
  border-left: 3px solid #0f766e;
  border-radius: 0 10px 10px 0;
  background: rgba(15, 23, 42, 0.85);
  color: #7c8c9f;
}

.preview :deep(hr) {
  margin: 1.5em 0;
  border: none;
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent,
    rgba(51, 65, 85, 0.75),
    transparent
  );
}

.preview :deep(pre) {
  margin: 1em 0;
  padding: 14px 16px;
  border-radius: 10px;
  background: #000000;
  border: 1px solid rgba(30, 41, 59, 0.95);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.02);
  overflow-x: auto;
}

.preview :deep(pre code) {
  background: transparent;
  padding: 0;
  font-size: 13px;
  line-height: 1.65;
  color: #94a3b8;
}

.preview :deep(code) {
  background: rgba(15, 23, 42, 0.95);
  border: 1px solid rgba(51, 65, 85, 0.8);
  border-radius: 6px;
  padding: 2px 8px;
  font-size: 0.88em;
  color: #7dd3fc;
}

.preview :deep(a) {
  color: #38bdf8;
  font-weight: 600;
  text-decoration: underline;
  text-decoration-color: rgba(56, 189, 248, 0.35);
  text-underline-offset: 3px;
  transition: color 0.15s ease;
}

.preview :deep(a:hover) {
  color: #7dd3fc;
}

/**
 * 后端返回的 HTML 可能含浅色 class / 内联样式，强制压暗以与预览区一致
 */
.preview.preview--from-server {
  color: #94a3b8 !important;
  background: #020617 !important;
}

.preview.preview--from-server :deep(*) {
  border-color: rgba(51, 65, 85, 0.85) !important;
}

.preview.preview--from-server :deep(p),
.preview.preview--from-server :deep(span),
.preview.preview--from-server :deep(div),
.preview.preview--from-server :deep(td),
.preview.preview--from-server :deep(th),
.preview.preview--from-server :deep(li) {
  color: #8696ac !important;
  background-color: transparent !important;
}

.preview.preview--from-server :deep(h1),
.preview.preview--from-server :deep(h2),
.preview.preview--from-server :deep(h3),
.preview.preview--from-server :deep(h4),
.preview.preview--from-server :deep(h5),
.preview.preview--from-server :deep(h6) {
  color: #cbd5e1 !important;
  background-color: transparent !important;
}

.preview.preview--from-server :deep(strong),
.preview.preview--from-server :deep(b) {
  color: #e2e8f0 !important;
  background-color: transparent !important;
}

.preview.preview--from-server :deep(em),
.preview.preview--from-server :deep(i) {
  color: #94a3b8 !important;
}

.preview.preview--from-server :deep(a) {
  color: #38bdf8 !important;
  background-color: transparent !important;
}

.preview.preview--from-server :deep(blockquote) {
  color: #7c8c9f !important;
  background: rgba(15, 23, 42, 0.9) !important;
  border-left-color: #0f766e !important;
}

.preview.preview--from-server :deep(pre) {
  color: #94a3b8 !important;
  background: #000000 !important;
  border-color: rgba(30, 41, 59, 0.95) !important;
}

.preview.preview--from-server :deep(code) {
  color: #7dd3fc !important;
  background: rgba(15, 23, 42, 0.95) !important;
  border-color: rgba(51, 65, 85, 0.85) !important;
}

.preview.preview--from-server :deep(pre code) {
  color: #94a3b8 !important;
  background: transparent !important;
  border: none !important;
}

.preview.preview--from-server :deep(table) {
  background: #020617 !important;
  border-collapse: collapse !important;
}

.preview.preview--from-server :deep(th),
.preview.preview--from-server :deep(td) {
  border: 1px solid rgba(51, 65, 85, 0.85) !important;
  background: rgba(2, 6, 23, 0.95) !important;
}

.preview.preview--from-server :deep(thead) {
  background: rgba(15, 23, 42, 0.95) !important;
}

.preview.preview--from-server :deep(hr) {
  background: linear-gradient(
    90deg,
    transparent,
    rgba(51, 65, 85, 0.75),
    transparent
  ) !important;
  border: none !important;
}

.preview.preview--from-server :deep(img) {
  opacity: 0.92;
  border-radius: 8px;
}

.error-tip {
  margin: 0;
  font-size: 12px;
  color: #fca5a5;
}

.btn {
  border-radius: 8px;
  border: 1px solid transparent;
  padding: 7px 11px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 700;
}

.btn.secondary {
  background: rgba(51, 65, 85, 0.75);
  border-color: rgba(148, 163, 184, 0.2);
  color: #e2e8f0;
}

.btn.danger {
  background: rgba(220, 38, 38, 0.16);
  border-color: rgba(248, 113, 113, 0.35);
  color: #fecaca;
}

.btn.accent {
  background: rgba(16, 185, 129, 0.22);
  border-color: rgba(52, 211, 153, 0.35);
  color: #a7f3d0;
}

.btn.accent:hover {
  background: rgba(16, 185, 129, 0.32);
  border-color: rgba(52, 211, 153, 0.5);
}

@media (max-width: 900px) {
  .panes {
    grid-template-columns: 1fr;
  }

  .pane {
    min-height: 360px;
  }
}
</style>
