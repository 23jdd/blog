<script setup>
import { computed, onBeforeUnmount, ref } from "vue"
import Item from "./Item.vue"
import { excerptFromMarkdown } from "@/api/blogApi.js"

const props = defineProps({
  drafts: {
    type: Array,
    default: () => [
      {
        id: 101,
        title: "Draft: Vue 组件通信笔记",
        excerpt: "记录 props / emit / v-model 在复杂页面中的搭配方式。",
        updatedAt: "2026-03-26 10:22",
        status: "编辑中",
        tag: "前端"
      },
      {
        id: 102,
        title: "Draft: 暗黑主题色彩系统",
        excerpt: "整理灰阶、强调色与状态色的映射关系，确保组件视觉统一。",
        updatedAt: "2026-03-25 21:08",
        status: "待发布",
        tag: "设计"
      },
      {
        id: 103,
        title: "Draft: 评论系统设计草稿",
        excerpt: "多级评论与折叠交互草稿。",
        updatedAt: "2026-03-24 15:00",
        status: "编辑中",
        tag: "后端"
      }
    ]
  }
})

const emit = defineEmits(["publish", "edit-draft"])

const selectedTag = ref("")

const tagOptions = computed(() => {
  const set = new Set()
  const list = props.drafts || []
  for (const d of list) {
    if (d?.tag) set.add(d.tag)
  }
  return Array.from(set).sort()
})

const filteredDrafts = computed(() => {
  const list = props.drafts || []
  if (!selectedTag.value) return list
  return list.filter((d) => d.tag === selectedTag.value)
})

function toItem(draft) {
  const desc = draft.excerpt || excerptFromMarkdown(draft.content || "")
  return {
    avatar: draft.cover || "",
    name: draft.title || "未命名草稿",
    desc: desc || "暂无内容",
    tag: `${draft.tag || "未分类"} · ${draft.status || "草稿"} · ${draft.updatedAt || ""}`
  }
}

function goEdit(draft) {
  emit("edit-draft", draft)
}

/** 发布弹窗：可选封面图 + 可选最终文章标签；父组件将走 POST /file/uploadArticle（字段 article）再 POST /articles.cover_url */
const publishModalDraft = ref(null)
const publishConvertImage = ref(null)
const publishImagePreviewUrl = ref("")
const publishFileInputRef = ref(null)
const publishTag = ref("未分类")
const publishTagOptions = ["前端", "设计", "后端", "工具", "未分类"]

function revokePublishPreview() {
  if (publishImagePreviewUrl.value) {
    URL.revokeObjectURL(publishImagePreviewUrl.value)
    publishImagePreviewUrl.value = ""
  }
}

function openPublishModal(draft) {
  revokePublishPreview()
  publishConvertImage.value = null
  publishModalDraft.value = draft
  publishTag.value = draft.tag || "未分类"
  if (publishFileInputRef.value) publishFileInputRef.value.value = ""
}

function closePublishModal() {
  revokePublishPreview()
  publishConvertImage.value = null
  publishModalDraft.value = null
  if (publishFileInputRef.value) publishFileInputRef.value.value = ""
}

function onPublishImageChange(ev) {
  const file = ev.target?.files?.[0]
  revokePublishPreview()
  publishConvertImage.value = file || null
  if (file && file.type.startsWith("image/")) {
    publishImagePreviewUrl.value = URL.createObjectURL(file)
  }
}

function triggerPublishFilePick() {
  publishFileInputRef.value?.click()
}

function confirmPublish() {
  const draft = publishModalDraft.value
  if (!draft) return
  emit("publish", {
    draft,
    convert_image: publishConvertImage.value,
    tag: publishTag.value || draft.tag || ""
  })
  closePublishModal()
}

onBeforeUnmount(() => {
  revokePublishPreview()
})
</script>

<template>
  <div class="draft-list-page">
    <section class="panel">
      <header class="panel-head">
        <h3 class="panel-title">草稿箱</h3>
        <div class="toolbar">
          <label class="tag-label" for="draft-tag-filter">标签</label>
          <select id="draft-tag-filter" v-model="selectedTag" class="tag-select">
            <option value="">全部</option>
            <option v-for="t in tagOptions" :key="t" :value="t">{{ t }}</option>
          </select>
        </div>
      </header>

      <div v-if="filteredDrafts.length" class="list">
        <div v-for="draft in filteredDrafts" :key="draft.id" class="draft-row">
          <div class="draft-item-wrap">
            <Item :item="toItem(draft)" />
          </div>
          <div class="draft-actions">
            <button type="button" class="edit-btn" @click="goEdit(draft)">编辑</button>
            <button
              type="button"
              class="publish-btn"
              :aria-label="`发布：${draft.title || '草稿'}`"
              @click="openPublishModal(draft)"
            >
              <span class="publish-btn-inner">
                <!-- 纸飞机图标：表示发布 / 发出 -->
                <svg class="publish-icon" viewBox="0 0 24 24" aria-hidden="true">
                  <path
                    fill="currentColor"
                    d="M2.01 21 23 12 2.01 3 2 10l15 2-15 2z"
                  />
                </svg>
                <span class="publish-label">发布</span>
              </span>
            </button>
          </div>
        </div>
      </div>
      <p v-else class="empty">暂无草稿（可切换标签或等待新草稿）</p>
    </section>

    <!-- 发布：可选封面图 → 上传后作为 CreateArticleRequest.cover_url -->
    <Teleport to="body">
      <div
        v-if="publishModalDraft"
        class="publish-modal-backdrop"
        role="dialog"
        aria-modal="true"
        aria-labelledby="publish-modal-title"
        @click.self="closePublishModal"
      >
        <div class="publish-modal">
          <h3 id="publish-modal-title" class="publish-modal-title">发布草稿</h3>
          <p class="publish-modal-draft-name">{{ publishModalDraft.title || "未命名草稿" }}</p>
          <p class="publish-modal-hint">
            可选封面图：将经 <code class="publish-field-name">/file/uploadArticle</code> 上传后写入文章的
            <code class="publish-field-name">cover_url</code>（与 swagger 一致）。
          </p>

          <div class="publish-tag-row">
            <label class="publish-tag-label" for="publish-tag">文章标签</label>
            <select
              id="publish-tag"
              v-model="publishTag"
              class="publish-tag-select"
            >
              <option v-for="t in publishTagOptions" :key="t" :value="t">
                {{ t }}
              </option>
            </select>
          </div>

          <input
            ref="publishFileInputRef"
            type="file"
            class="publish-file-input"
            accept="image/*"
            @change="onPublishImageChange"
          />

          <div class="publish-image-row">
            <button type="button" class="publish-pick-btn" @click="triggerPublishFilePick">
              选择图片
            </button>
            <span v-if="publishConvertImage" class="publish-file-name" :title="publishConvertImage.name">
              {{ publishConvertImage.name }}
            </span>
            <span v-else class="publish-file-placeholder">未选择</span>
          </div>

          <div v-if="publishImagePreviewUrl" class="publish-preview-wrap">
            <img :src="publishImagePreviewUrl" alt="预览" class="publish-preview-img" />
          </div>

          <div class="publish-modal-actions">
            <button type="button" class="publish-cancel-btn" @click="closePublishModal">取消</button>
            <button type="button" class="publish-confirm-btn" @click="confirmPublish">确认发布</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.draft-list-page {
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
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 6px 6px 14px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.12);
  margin-bottom: 12px;
}

.panel-title {
  margin: 0;
  font-size: 16px;
  font-weight: 800;
  color: #e5e7eb;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tag-label {
  font-size: 13px;
  color: #94a3b8;
  font-weight: 600;
}

.tag-select {
  min-width: 120px;
  padding: 8px 10px;
  border-radius: 8px;
  border: 1px solid rgba(148, 163, 184, 0.28);
  background: rgba(15, 23, 42, 0.75);
  color: #e2e8f0;
  font-size: 13px;
  outline: none;
  cursor: pointer;
}

.tag-select:focus {
  border-color: rgba(96, 165, 250, 0.45);
}

.list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.draft-row {
  display: flex;
  align-items: stretch;
  gap: 10px;
}

.draft-item-wrap {
  flex: 1;
  min-width: 0;
}

.draft-actions {
  flex-shrink: 0;
  align-self: center;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.edit-btn {
  padding: 10px 16px;
  border-radius: 10px;
  border: 1px solid rgba(96, 165, 250, 0.4);
  background: rgba(59, 130, 246, 0.18);
  color: #93c5fd;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  transition: background-color 0.15s ease, border-color 0.15s ease;
  white-space: nowrap;
}

.edit-btn:hover {
  background: rgba(59, 130, 246, 0.28);
  border-color: rgba(96, 165, 250, 0.55);
}

.publish-btn {
  position: relative;
  overflow: hidden;
  padding: 10px 16px;
  border-radius: 10px;
  border: 1px solid rgba(52, 211, 153, 0.35);
  background: linear-gradient(
    145deg,
    rgba(16, 185, 129, 0.22) 0%,
    rgba(5, 150, 105, 0.14) 100%
  );
  color: #a7f3d0;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
  isolation: isolate;
  transition:
    background 0.2s ease,
    border-color 0.2s ease,
    box-shadow 0.25s ease,
    transform 0.12s ease;
  box-shadow: 0 0 0 0 rgba(52, 211, 153, 0);
}

/* 悬停高光扫过（shine） */
.publish-btn::before {
  content: "";
  position: absolute;
  inset: 0;
  left: -100%;
  width: 60%;
  background: linear-gradient(
    100deg,
    transparent 0%,
    rgba(255, 255, 255, 0.18) 45%,
    rgba(255, 255, 255, 0.06) 55%,
    transparent 100%
  );
  transform: skewX(-18deg);
  opacity: 0;
  transition: left 0.55s ease, opacity 0.2s ease;
  pointer-events: none;
  z-index: 0;
}

.publish-btn:hover::before {
  left: 120%;
  opacity: 1;
}

.publish-btn-inner {
  position: relative;
  z-index: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.publish-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  filter: drop-shadow(0 0 6px rgba(52, 211, 153, 0.35));
  transition: transform 0.2s ease, filter 0.2s ease;
}

.publish-btn:hover .publish-icon {
  transform: translate(2px, -1px) rotate(-6deg);
  filter: drop-shadow(0 0 10px rgba(110, 231, 183, 0.55));
}

.publish-btn:hover {
  border-color: rgba(110, 231, 183, 0.65);
  background: linear-gradient(
    145deg,
    rgba(16, 185, 129, 0.35) 0%,
    rgba(5, 150, 105, 0.22) 100%
  );
  box-shadow:
    0 0 20px rgba(52, 211, 153, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.publish-btn:active {
  transform: scale(0.97);
}

.publish-btn:focus-visible {
  outline: 2px solid rgba(110, 231, 183, 0.85);
  outline-offset: 2px;
}

.empty {
  margin: 16px 6px;
  color: #94a3b8;
  font-size: 14px;
}

.publish-modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: rgba(2, 6, 23, 0.72);
  backdrop-filter: blur(6px);
}

.publish-modal {
  width: min(420px, 100%);
  padding: 22px;
  border-radius: 16px;
  background: rgba(15, 23, 42, 0.96);
  border: 1px solid rgba(148, 163, 184, 0.22);
  box-shadow: 0 24px 48px rgba(0, 0, 0, 0.45);
}

.publish-modal-title {
  margin: 0 0 8px;
  font-size: 17px;
  font-weight: 800;
  color: #f1f5f9;
}

.publish-modal-draft-name {
  margin: 0 0 12px;
  font-size: 14px;
  color: #94a3b8;
  line-height: 1.45;
  word-break: break-word;
}

.publish-modal-hint {
  margin: 0 0 14px;
  font-size: 13px;
  color: #cbd5e1;
  line-height: 1.5;
}

.publish-tag-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin: 0 0 12px;
}

.publish-tag-label {
  font-size: 13px;
  font-weight: 600;
  color: #94a3b8;
}

.publish-tag-select {
  min-width: 140px;
  padding: 8px 10px;
  border-radius: 8px;
  border: 1px solid rgba(148, 163, 184, 0.32);
  background: rgba(15, 23, 42, 0.85);
  color: #e2e8f0;
  font-size: 13px;
  outline: none;
  cursor: pointer;
}

.publish-tag-select:focus {
  border-color: rgba(96, 165, 250, 0.6);
}

.publish-field-name {
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 6px;
  background: rgba(51, 65, 85, 0.85);
  color: #a7f3d0;
}

.publish-file-input {
  position: absolute;
  width: 0;
  height: 0;
  opacity: 0;
  pointer-events: none;
}

.publish-image-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.publish-pick-btn {
  padding: 8px 14px;
  border-radius: 10px;
  border: 1px solid rgba(52, 211, 153, 0.4);
  background: rgba(16, 185, 129, 0.2);
  color: #a7f3d0;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}

.publish-pick-btn:hover {
  background: rgba(16, 185, 129, 0.32);
}

.publish-file-name {
  font-size: 12px;
  color: #e2e8f0;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.publish-file-placeholder {
  font-size: 12px;
  color: #64748b;
}

.publish-preview-wrap {
  margin-bottom: 16px;
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid rgba(148, 163, 184, 0.2);
  background: rgba(2, 6, 23, 0.5);
}

.publish-preview-img {
  display: block;
  width: 100%;
  max-height: 200px;
  object-fit: contain;
}

.publish-modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 4px;
}

.publish-cancel-btn {
  padding: 9px 16px;
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.35);
  background: rgba(30, 41, 59, 0.8);
  color: #e2e8f0;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.publish-confirm-btn {
  padding: 9px 16px;
  border-radius: 10px;
  border: 1px solid rgba(110, 231, 183, 0.45);
  background: linear-gradient(145deg, rgba(16, 185, 129, 0.35), rgba(5, 150, 105, 0.25));
  color: #ecfdf5;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}

.publish-confirm-btn:hover {
  filter: brightness(1.06);
}

@media (max-width: 560px) {
  .draft-row {
    flex-direction: column;
  }

  .draft-actions {
    flex-direction: row;
    align-self: stretch;
  }

  .edit-btn,
  .publish-btn {
    flex: 1;
  }
}
</style>
