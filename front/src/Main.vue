<template>
  <main class="layout">
    <Option
      :items="menus"
      :active-key="activeMenu"
      title="GO"
      @select="handleSelectMenu"
    />

    <section class="content">
      <UserInfo :user="userInfo" @update:user="handleUserUpdate" @save:personImage="handleSavePersonImage" v-if="activeMenu === 'info'" />
      <ItemList
        v-else-if="activeMenu === '推荐'"
        :articles="feedArticles"
        :loading="feedLoading"
        :error="feedError"
        @open-read="handleOpenRead"
      />
      <Draft
        v-else-if="activeMenu === '草稿'"
        :open-from-draft-box="openDraftFromBox"
        @save-to-draft-list="handleSaveToDraftList"
      />
      <Read v-else-if="activeMenu === '阅读'" :article-id="readArticleId" />
      <DraftList
        v-else-if="activeMenu==='草稿箱'"
        :drafts="draftBox"
        @publish="handlePublishDraft"
        @edit-draft="handleEditDraftFromList"
      />
      <ArticleList
        v-else-if="activeMenu === '我的文章'"
        :articles="authorArticles"
        :loading="authorLoading"
      />
      <Search v-else-if="activeMenu === '搜索'" @select="handleOpenRead" />
      <Like v-else-if="activeMenu === '我的点赞'" @open-read="handleOpenRead" />
      <Colletc v-else-if="activeMenu === '我的收藏'" @open-read="handleOpenRead" />
     </section>
  </main>
</template>

<script setup>
import Draft from "@/components/Draft.vue"
import { ref, onMounted, watch } from "vue"
import Option from "@/components/Option.vue"
import UserInfo from "@/components/UserInfo.vue"
import ItemList from "@/components/ItemList.vue"
import Read from "./components/Read.vue"
import DraftList from "./components/DraftList.vue"
import ArticleList from "./components/ArticleList.vue"
import Search from "./components/Search.vue"
import Like from "./components/Like.vue"
import Colletc from "./components/Colletc.vue"
import { API_BASE } from "@/api/config.js"
import { getAccessToken } from "@/api/http.js"
import {
  serverDraftToUiAsync,
  excerptFromMarkdown,
  fetchDraftsList,
  createDraft,
  updateDraft,
  deleteDraft,
  createArticle,
  uploadArticleFile,
  uploadResultToUrl,
  resolveMediaUrl,
  fetchUserInfo,
  pickUserImage,
  pickScalar,
  articleModelToCard,
  normalizeFeedPayload,
  fetchArticleById,
  fetchHotArticles,
  fetchFeed,
  fetchArticlesByAuthor
} from "@/api/blogApi.js"
const menus = [
  { key: "info", label: "Info" },
  { key: "推荐", label: "推荐" },
  { key: "草稿", label: "草稿" },
  { key: "我的文章", label: "我的文章" },
  {key:"草稿箱",label:"草稿箱"},
  { key: "搜索", label: "搜索" },
  { key: "我的点赞", label: "我的点赞" },
  { key: "我的收藏", label: "我的收藏" },
]

const activeMenu = ref("info")
const Baseurl = API_BASE

/** 从草稿箱「编辑」带入 Draft；含 _ts 供子组件 watch 触发加载 */
const openDraftFromBox = ref(null)

/** 与 swagger GET /drafts 同步的本地列表展示结构 */
const draftBox = ref([])

const currentUserId = ref(null)
const readArticleId = ref(null)

const feedArticles = ref([])
const feedLoading = ref(false)
const feedError = ref("")

const authorArticles = ref([])
const authorLoading = ref(false)

const userInfo = ref({
  avatar: "",
  nickname: "",
  email: "",
  age: "",
  gender: "",
  bio: "专注前端体验与界面设计，喜欢做干净、舒服、可维护的组件。",
  createdAt: ""
})

onMounted(async () => {
  await GetUserInfo()
  await loadDraftsFromServer()
  await loadFeedArticles()
})

watch(activeMenu, (k) => {
  if (k === "草稿箱") loadDraftsFromServer()
  if (k === "推荐") loadFeedArticles()
  if (k === "我的文章") loadAuthorArticles()
  if (k !== "阅读") readArticleId.value = null
})
function handleSelectMenu(item) {
  if (item.key === "草稿") {
    openDraftFromBox.value = null
  }
  activeMenu.value = item.key
}

function handleEditDraftFromList(draft) {
  openDraftFromBox.value = { ...draft, _ts: Date.now() }
  activeMenu.value = "草稿"
}

function handleOpenRead(article) {
  const id = article?.id
  readArticleId.value = id != null ? id : null
  activeMenu.value = "阅读"
}

async function loadDraftsFromServer() {
  if (!getAccessToken()) return
  try {
    const list = await fetchDraftsList()
    const arr = Array.isArray(list) ? list : []
    draftBox.value = await Promise.all(arr.map((d) => serverDraftToUiAsync(d)))
  } catch (e) {
    console.warn("草稿列表同步失败", e?.message || e)
  }
}

async function loadFeedArticles() {
  feedLoading.value = true
  feedError.value = ""
  try {
    const token = getAccessToken()
    let raw
    if (token) {
      try {
        raw = await fetchFeed(1, 24)
      } catch {
        raw = await fetchHotArticles(24)
      }
    } else {
      raw = await fetchHotArticles(24)
    }
    let items = normalizeFeedPayload(raw).slice(0, 18)
    const cards = await Promise.all(
      items.map(async (it) => {
        if (it && typeof it === "object" && it.title) return articleModelToCard(it)
        const id = it?.id ?? it
        if (id == null || Number.isNaN(Number(id))) return null
        try {
          const a = await fetchArticleById(Number(id))
          return articleModelToCard(a)
        } catch {
          return {
            id: Number(id),
            title: `文章 #${id}`,
            summary: "",
            category: "文章",
            author: "",
            date: "",
            readTime: "",
            cover: ""
          }
        }
      })
    )
    feedArticles.value = cards.filter(Boolean)
  } catch (e) {
    feedError.value = e?.message || "加载推荐失败"
    feedArticles.value = []
  } finally {
    feedLoading.value = false
  }
}

async function loadAuthorArticles() {
  authorLoading.value = true
  try {
    if (!currentUserId.value) {
      authorArticles.value = []
      return
    }
    const raw = await fetchArticlesByAuthor(currentUserId.value, {
      page: 1,
      pageSize: 50
    })
    authorArticles.value = raw.map((a) => articleModelToCard(a)).filter(Boolean)
  } catch (e) {
    console.warn("我的文章加载失败", e?.message || e)
    authorArticles.value = []
  } finally {
    authorLoading.value = false
  }
}

async function GetUserInfo() {
  const token = getAccessToken()
  if (!token) return
  try {
    const r = await fetchUserInfo()
    currentUserId.value = r?.id ?? null
    userInfo.value = {
      avatar: pickUserImage(r?.image),
      nickname: r?.username ?? "",
      email: r?.username ?? "",
      age: pickScalar(r?.age) || "",
      gender: pickScalar(r?.gender) || "",
      bio: userInfo.value.bio,
      createdAt: r?.created_at || ""
    }
  } catch (e) {
    console.warn("用户信息失败", e?.message || e)
  }
}
async function UpdateUserInfo(age,gender) {
    let token = window.localStorage.getItem("accessToken")
    if (token) {
        let r = await fetch(Baseurl + "/user/update", {
            body: `age=${age}&gender=${gender}`,
            method: "POST",
            headers: {
                Authorization: `Bearer ${token}`,
                "Content-Type": "application/x-www-form-urlencoded",
            }

        })
        let data = await r.json().catch(() => ({}))
        if (data){
            console.log(data)
        }
    }
}
async function SavePersonImage(formData,user) {
    const token = window.localStorage.getItem("accessToken")
    if (!token) {
        console.warn("未登录，无法上传头像")
        return
    }
    try {
        // FormData 上传时不要设置 Content-Type，由浏览器自动带 boundary
        const r = await fetch(Baseurl + "/file/setPersonImage", {
            method: "POST",
            body: formData,
            headers: {
                Authorization: `Bearer ${token}`
            }
        })
        const res = await r.json().catch(() => ({}))
        if (!r.ok) {
            console.error("上传头像失败", res?.message || r.status)
            return
        }
        console.log("上传头像成功", res)
        const url = res?.data?.url ?? res?.url
        userInfo.value.avatar = url ? (url.startsWith("http") ? url : Baseurl + url) : userInfo.value.avatar
    } catch (e) {
        // 后端未启动、跨域、断网等会进这里，避免 Uncaught (in promise)
        console.error("上传头像请求失败（请检查后端是否运行在 " + Baseurl + "）", e)
    }
}
function handleUserUpdate(nextUser) {
  userInfo.value = { ...nextUser }
  UpdateUserInfo(nextUser.age,nextUser.gender)
}
function handleSavePersonImage(formData,user) {
   SavePersonImage(formData).catch(() => {})
}

function formatDraftTime() {
  const d = new Date()
  const p = (n) => String(n).padStart(2, "0")
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}

/** 存入草稿箱：已登录走 POST/PUT /drafts；未登录仅本地列表 */
async function handleSaveToDraftList(entry) {
  const title = entry.title?.trim() || "未命名草稿"
  const content = entry.content ?? ""
  const body = { title, content }
  const token = getAccessToken()

  if (!token) {
    const id = entry.id
    const idx =
      id != null
        ? draftBox.value.findIndex((d) => String(d.id) === String(id))
        : -1
    if (idx !== -1) {
      const prev = draftBox.value[idx]
      draftBox.value.splice(idx, 1, {
        ...prev,
        title,
        excerpt: excerptFromMarkdown(content),
        updatedAt: formatDraftTime(),
        status: prev.status || "编辑中",
        tag: entry.tag || "未分类",
        content
      })
    } else {
      draftBox.value.unshift({
        id: Date.now(),
        title,
        excerpt: excerptFromMarkdown(content),
        updatedAt: formatDraftTime(),
        status: "编辑中",
        tag: entry.tag || "未分类",
        content
      })
    }
    return
  }

  try {
    if (entry.id != null) {
      try {
        await updateDraft(entry.id, body)
      } catch (e) {
        if (e.status === 404) await createDraft(body)
        else throw e
      }
    } else {
      await createDraft(body)
    }
    await loadDraftsFromServer()
  } catch (e) {
    window.alert(e?.message || "保存草稿失败（/drafts）")
  }
}

/**
 * 发布草稿：swagger 为 POST /articles（CreateArticleRequest，含 draft_id、cover_url 等）
 * 封面图先 POST /file/uploadArticle（表单字段 article），再将返回 URL 写入 cover_url
 */
async function handlePublishDraft(payload) {
  const draft = payload?.draft != null ? payload.draft : payload
  const convertImage =
    payload && typeof payload === "object" && "convert_image" in payload
      ? payload.convert_image
      : null

  const token = getAccessToken()
  if (!token) {
    draftBox.value = draftBox.value.filter((d) => d.id !== draft.id)
    return
  }

  try {
    let cover_url = ""
    if (convertImage instanceof File) {
      const up = await uploadArticleFile(convertImage)
      const raw = uploadResultToUrl(up)
      cover_url = raw
        ? /^https?:\/\//i.test(raw)
          ? raw
          : resolveMediaUrl(raw)
        : ""
    }
    const finalTag =
      payload && typeof payload === "object" && "tag" in payload
        ? payload.tag || draft.tag || ""
        : draft.tag || ""
    await createArticle({
      draft_id: draft.id,
      title: draft.title || "未命名",
      tags: finalTag,
      status: "published",
      ...(cover_url ? { cover_url } : {})
    })
    try {
      await deleteDraft(draft.id)
    } catch {
      /* 服务端可能已随发布删除草稿 */
    }
    await loadDraftsFromServer()
  } catch (e) {
    console.error(e)
    window.alert(e?.message || "发布失败（POST /articles）")
  }
}
</script>

<style scoped>
.layout {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 260px 1fr;
  background: linear-gradient(180deg, #090b0d 0%, #111827 100%);
}

.content {
  padding: 20px 24px;
  overflow: auto;
}

@media (max-width: 860px) {
  .layout {
    grid-template-columns: 1fr;
  }

  .content {
    padding: 16px;
  }
}
</style>