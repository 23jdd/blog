import { API_BASE } from "./config.js"
import {
  apiFetch,
  unwrapData,
  authHeaders,
  getAccessToken,
  parseJsonSafe
} from "./http.js"

/** 拼完整资源 URL（相对路径补全为 API 根） */
export function resolveMediaUrl(path) {
  if (!path) return ""
  if (/^https?:\/\//i.test(path)) return path
  return `${API_BASE.replace(/\/$/, "")}/${String(path).replace(/^\//, "")}`
}

/** Go sql.Null* 或普通 JSON 字段 */
export function pickScalar(field) {
  if (field == null) return ""
  if (typeof field === "object" && "Valid" in field) {
    return field.Valid ? field.String ?? field.Int32 ?? field.Int64 ?? "" : ""
  }
  return field
}

export function pickUserImage(imageField) {
  const s = pickScalar(imageField)
  if (!s) return ""
  return resolveMediaUrl(s)
}

/** GET /drafts → 草稿数组 */
export async function fetchDraftsList() {
  const json = await apiFetch("/drafts", { method: "GET" })
  const data = unwrapData(json)
  if (Array.isArray(data)) return data
  if (Array.isArray(data?.list)) return data.list
  if (Array.isArray(data?.drafts)) return data.drafts
  return []
}

/** POST /drafts */
export async function createDraft(body) {
  const json = await apiFetch("/drafts", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body)
  })
  return unwrapData(json)
}

/** PUT /drafts/{id} */
export async function updateDraft(id, body) {
  const json = await apiFetch(`/drafts/${encodeURIComponent(id)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body)
  })
  return unwrapData(json)
}

/** DELETE /drafts/{id} */
export async function deleteDraft(id) {
  const json = await apiFetch(`/drafts/${encodeURIComponent(id)}`, {
    method: "DELETE"
  })
  return unwrapData(json)
}

/** POST /articles（发布草稿见 swagger types.CreateArticleRequest，含 draft_id） */
export async function createArticle(body) {
  const json = await apiFetch("/articles", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body)
  })
  return unwrapData(json)
}

/** POST /file/uploadArticle，表单字段名 article */
export async function uploadArticleFile(file) {
  const fd = new FormData()
  fd.append("article", file, file.name)
  const token = getAccessToken()
  const r = await fetch(`${API_BASE}/file/uploadArticle`, {
    method: "POST",
    headers: token ? { Authorization: `Bearer ${token}` } : {},
    body: fd
  })
  const json = await parseJsonSafe(r)
  if (!r.ok) {
    throw new Error(json?.message || `上传失败（HTTP ${r.status}）`)
  }
  return unwrapData(json)
}


/** 从上传结果中取可写入 cover_url 的路径 */
export function uploadResultToUrl(result) {
  if (!result || typeof result !== "object") return ""
  const u = result.url ?? result.path ?? result.file_url
  if (typeof u === "string") return u
  if (result.data && typeof result.data === "object") {
    const d = result.data
    return d.url ?? d.path ?? ""
  }
  return ""
}

/** GET /articles/{id} */
export async function fetchArticleById(id) {
  const json = await apiFetch(`/articles/${encodeURIComponent(id)}`, {
    method: "GET"
  })
  return unwrapData(json)
}

/** GET /articles/{id}/stats（访问量、点赞数等，字段名依后端为准） */
export async function fetchArticleStats(id) {
  const json = await apiFetch(`/articles/${encodeURIComponent(id)}/stats`, {
    method: "GET"
  })
  return unwrapData(json)
}

/** POST /articles/{id}/likes — 返回 types.LikeActionResponse { liked } */
export async function likeArticle(id) {
  const json = await apiFetch(`/articles/${encodeURIComponent(id)}/likes`, {
    method: "POST"
  })
  return unwrapData(json)
}

/** DELETE /articles/{id}/likes */
export async function unlikeArticle(id) {
  const json = await apiFetch(`/articles/${encodeURIComponent(id)}/likes`, {
    method: "DELETE"
  })
  return unwrapData(json)
}

/** POST /articles/{id}/collections — 收藏 */
export async function collectArticle(id) {
  const json = await apiFetch(`/articles/${encodeURIComponent(id)}/collections`, {
    method: "POST"
  })
  return unwrapData(json)
}

/** DELETE /articles/{id}/collections — 取消收藏 */
export async function uncollectArticle(id) {
  const json = await apiFetch(`/articles/${encodeURIComponent(id)}/collections`, {
    method: "DELETE"
  })
  return unwrapData(json)
}

/** GET /articles/{id}/comments */
export async function fetchArticleComments(id, page = 1, pageSize = 200) {
  const q = new URLSearchParams()
  if (page != null) q.set("page", String(page))
  if (pageSize != null) q.set("pageSize", String(pageSize))
  const json = await apiFetch(
    `/articles/${encodeURIComponent(id)}/comments?${q.toString()}`,
    { method: "GET" }
  )
  const data = unwrapData(json)
  return Array.isArray(data) ? data : []
}

/** POST /articles/{id}/comments */
export async function createArticleComment(id, payload) {
  const json = await apiFetch(`/articles/${encodeURIComponent(id)}/comments`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload)
  })
  return unwrapData(json)
}

/** POST /interactions/follow/{targetID} — 关注用户 */
export async function followUser(targetID) {
  const json = await apiFetch(
    `/interactions/follow/${encodeURIComponent(targetID)}`,
    { method: "POST" }
  )
  return unwrapData(json)
}

/** DELETE /interactions/follow/{targetID} — 取消关注用户 */
export async function unfollowUser(targetID) {
  const json = await apiFetch(
    `/interactions/follow/${encodeURIComponent(targetID)}`,
    { method: "DELETE" }
  )
  return unwrapData(json)
}

/** GET /interactions/my-collections — 用于判断是否已收藏 */
export async function fetchMyCollections(page = 1, pageSize = 100) {
  const q = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize)
  })
  const json = await apiFetch(`/interactions/my-collections?${q.toString()}`, {
    method: "GET"
  })
  const data = unwrapData(json)
  return Array.isArray(data) ? data : []
}

/**
 * GET /interactions/my-likes（swagger 未列出时常见约定；若无此路由会 404）
 */
export async function fetchMyLikes(page = 1, pageSize = 100) {
  const q = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize)
  })
  const json = await apiFetch(`/interactions/my-likes?${q.toString()}`, {
    method: "GET"
  })
  const data = unwrapData(json)
  return Array.isArray(data) ? data : []
}

/** 收藏/点赞列表行 → 列表卡片（兼容 model.Collect 等字段名） */
export function interactionRowToArticle(row, kindLabel) {
  if (!row || typeof row !== "object") return null
  const id = Number(row.articleID ?? row.article_id ?? row.articleId)
  if (!id || Number.isNaN(id)) return null
  const title =
    row.articleTitle ??
    row.article_title ??
    row.title ??
    `文章 #${id}`
  const time =
    row.createTime ??
    row.create_time ??
    row.updateTime ??
    row.update_time ??
    row.likedAt ??
    row.liked_at ??
    ""
  return {
    id,
    title,
    summary: time ? `${kindLabel} · ${time}` : kindLabel,
    category: kindLabel,
    date: typeof time === "string" ? time : "",
    cover: ""
  }
}

/** GET /articles/author/{authid} */
export async function fetchArticlesByAuthor(authId, params = {}) {
  const q = new URLSearchParams()
  if (params.page != null) q.set("page", String(params.page))
  if (params.pageSize != null) q.set("pageSize", String(params.pageSize))
  const qs = q.toString()
  const path = `/articles/author/${encodeURIComponent(authId)}${qs ? `?${qs}` : ""}`
  const json = await apiFetch(path, { method: "GET" })
  const data = unwrapData(json)
  return Array.isArray(data) ? data : []
}

/** GET /articles/search */
export async function searchArticlesRemote(keyword, page = 1, pageSize = 10) {
  const q = new URLSearchParams({
    keyword: keyword || "",
    page: String(page),
    pageSize: String(pageSize)
  })
  const json = await apiFetch(`/articles/search?${q.toString()}`, {
    method: "GET"
  })
  const data = unwrapData(json)
  return Array.isArray(data) ? data : []
}

/** GET /articles/hot */
export async function fetchHotArticles(limit = 20) {
  const q = limit != null ? `?limit=${encodeURIComponent(String(limit))}` : ""
  const json = await apiFetch(`/articles/hot${q}`, { method: "GET" })
  return unwrapData(json)
}

/** GET /interactions/feed */
export async function fetchFeed(page = 1, pageSize = 20) {
  const q = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize)
  })
  const json = await apiFetch(`/interactions/feed?${q.toString()}`, {
    method: "GET"
  })
  return unwrapData(json)
}

/** POST /markdown：正文转 HTML（consumes 在 swagger 为 string body） */
export async function renderMarkdown(markdownText) {
  const json = await apiFetch("/markdown", {
    method: "POST",
    headers: { "Content-Type": "text/plain; charset=utf-8" },
    body: markdownText ?? ""
  })
  const data = unwrapData(json)
  if (typeof data === "string") return data
  return data?.html ?? data?.HTML ?? ""
}

/** GET /user/info */
export async function fetchUserInfo() {
  const json = await apiFetch("/user/info", { method: "GET" })
  return unwrapData(json)
}

/** 拉取远程 markdown 文本（article.content 常为文件 URL） */
export async function fetchTextUrl(url) {
  const full = resolveMediaUrl(url)
  const r = await fetch(full, { headers: authHeaders() })
  if (!r.ok) throw new Error(`加载正文失败（HTTP ${r.status}）`)
  return r.text()
}

export function excerptFromMarkdown(text) {
  const t = (text || "").replace(/\s+/g, " ").trim()
  if (!t) return "暂无正文"
  return t.length > 120 ? `${t.slice(0, 120)}…` : t
}

export function formatDraftTimeLabel(isoOrStr) {
  if (!isoOrStr) return ""
  try {
    const d = new Date(isoOrStr)
    if (!Number.isNaN(d.getTime())) {
      const p = (n) => String(n).padStart(2, "0")
      return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(
        d.getHours()
      )}:${p(d.getMinutes())}`
    }
  } catch {
    /* ignore */
  }
  return String(isoOrStr)
}

function looksLikeRemoteTextPath(v) {
  if (!v) return false
  const s = String(v).trim()
  if (!s) return false
  if (/^https?:\/\//i.test(s)) return true
  if (s.startsWith("/")) return true
  if (/\.(md|markdown|txt)(\?.*)?$/i.test(s)) return true
  if (s.includes("/file/getFile")) return true
  return false
}

/** 服务端草稿项 → 本地 DraftList / 编辑器使用的结构（自动解析 content_url） */
export async function serverDraftToUiAsync(d) {
  const rawContent = d?.content != null ? String(d.content) : ""
  const rawContentUrl = d?.content_url != null ? String(d.content_url) : ""

  let content = rawContent
  let contentUrl = ""
  let excerptSource = ""

  // 优先按 content_url；若没有且 content 像路径，也按 URL 拉取正文
  const candidateUrl = rawContentUrl || (looksLikeRemoteTextPath(rawContent) ? rawContent : "")
  if (candidateUrl) {
    contentUrl = resolveMediaUrl(candidateUrl)
    try {
      content = await fetchTextUrl(candidateUrl)
      excerptSource = content
    } catch {
      // 拉取失败时保留原内容（通常是 URL 字符串），避免数据丢失
      content = rawContent || candidateUrl
      // 注意：URL 不是正文，不用于生成摘要
      excerptSource = ""
    }
  } else {
    excerptSource = content
  }

  return {
    id: d.id,
    title: d.title || "未命名草稿",
    excerpt: excerptFromMarkdown(excerptSource),
    content,
    contentUrl,
    updatedAt: formatDraftTimeLabel(d.update_time || d.updated_at),
    tag: "未分类",
    status: typeof d.status === "string" ? d.status : "编辑中"
  }
}

/** model.Article → 列表卡片 */
export function articleModelToCard(a) {
  if (!a || typeof a !== "object") return null
  const tags = a.tags != null ? String(a.tags) : ""
  const firstTag = tags.split(",")[0]?.trim()
  return {
    id: a.id,
    title: a.title || "未命名",
    summary: tags || excerptFromMarkdown(String(a.content || "")),
    category: firstTag || (a.category_id != null ? `分类 ${a.category_id}` : "文章"),
    author: "",
    date: formatDraftTimeLabel(a.create_time || a.update_time),
    readTime: "",
    cover: resolveMediaUrl(a.cover_url || ""),
    _raw: a
  }
}

/** 解析 /articles/hot、/interactions/feed 等松散结构 */
export function normalizeFeedPayload(data) {
  if (data == null) return []
  if (Array.isArray(data)) {
    if (data.length === 0) return []
    if (typeof data[0] === "number") return data.map((id) => ({ id }))
    return data
  }
  if (Array.isArray(data.items)) return normalizeFeedPayload(data.items)
  if (Array.isArray(data.articles)) return normalizeFeedPayload(data.articles)
  if (Array.isArray(data.article_ids))
    return data.article_ids.map((id) => ({ id }))
  if (Array.isArray(data.ids)) return data.ids.map((id) => ({ id }))
  return []
}
