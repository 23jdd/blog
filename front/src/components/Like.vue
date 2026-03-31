<script setup>
import { onMounted, ref } from "vue"
import Item from "./Item.vue"
import { getAccessToken } from "@/api/http.js"
import {
  fetchFeed,
  fetchHotArticles,
  normalizeFeedPayload,
  fetchArticleById,
  articleModelToCard,
  fetchArticleStats
} from "@/api/blogApi.js"

const emit = defineEmits(["open-read"])

const articles = ref([])
const loading = ref(false)
const error = ref("")
const needLogin = ref(false)

async function loadCandidateArticleIds() {
  // swagger 没有「我的点赞列表」接口，这里基于 feed/hot + stats.liked 近似得到我的点赞
  let raw
  try {
    raw = await fetchFeed(1, 80)
  } catch {
    raw = await fetchHotArticles(120)
  }
  const items = normalizeFeedPayload(raw)
  const ids = []
  for (const it of items) {
    const id = Number(it?.id ?? it)
    if (!Number.isNaN(id) && id > 0 && !ids.includes(id)) ids.push(id)
    if (ids.length >= 80) break
  }
  return ids
}

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
    const ids = await loadCandidateArticleIds()
    const likedIds = []
    for (const id of ids) {
      try {
        const stats = await fetchArticleStats(id)
        const liked = stats?.liked
        if (liked === true) likedIds.push(id)
      } catch {
        // 单个文章统计失败不影响整体
      }
      if (likedIds.length >= 100) break
    }

    const cards = await Promise.all(
      likedIds.map(async (id) => {
        try {
          const a = await fetchArticleById(id)
          return articleModelToCard(a)
        } catch {
          return {
            id,
            title: `文章 #${id}`,
            summary: "我的点赞",
            category: "点赞",
            date: "",
            cover: ""
          }
        }
      })
    )
    articles.value = cards.filter(Boolean)
  } catch (e) {
    error.value = e?.message || "加载点赞列表失败"
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
    desc: a.summary || "我的点赞",
    tag: a.date || "点赞"
  }
}

function onSelectItem(it) {
  const id = it?.id
  if (id != null) emit("open-read", { id })
}
</script>

<template>
  <section class="my-panel">
    <header class="head">
      <h2 class="title">我的点赞</h2>
      <button type="button" class="refresh" :disabled="loading" @click="load">
        {{ loading ? "加载中…" : "刷新" }}
      </button>
    </header>

    <p v-if="needLogin" class="hint warn">请先登录后查看点赞列表。</p>
    <p v-else-if="error" class="hint error">{{ error }}</p>
    <p v-else-if="loading" class="hint">加载中…</p>
    <p v-else-if="!articles.length" class="hint">
      暂无点赞记录（基于 swagger 现有接口近似计算：feed/hot + article stats）。
    </p>

    <div v-else class="list">
      <Item
        v-for="a in articles"
        :key="a.id"
        :item="toItem(a)"
        @click="onSelectItem"
      />
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
  border: 1px solid rgba(248, 113, 113, 0.35);
  background: rgba(127, 29, 29, 0.25);
  color: #fecaca;
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
  line-height: 1.5;
}

.hint.warn {
  color: #fcd34d;
}

.list {
  display: grid;
  gap: 12px;
}
</style>
