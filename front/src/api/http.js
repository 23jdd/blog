import { API_BASE } from "./config.js"

export function getAccessToken() {
  return window.localStorage.getItem("accessToken") || ""
}

function getRefreshToken() {
  return window.localStorage.getItem("refreshToken") || ""
}

function setTokens(accessToken, refreshToken) {
  if (accessToken) {
    window.localStorage.setItem("accessToken", accessToken)
  }
  if (refreshToken) {
    window.localStorage.setItem("refreshToken", refreshToken)
  }
}

export function authHeaders(extra = {}) {
  const token = getAccessToken()
  const h = { ...extra }
  if (token) h.Authorization = `Bearer ${token}`
  return h
}

/** 解析统一响应：优先取 data，否则返回整包 */
export function unwrapData(json) {
  if (json == null || typeof json !== "object") return json
  if ("data" in json && json.data !== undefined) return json.data
  return json
}

export async function parseJsonSafe(r) {
  const text = await r.text()
  if (!text) return {}
  try {
    return JSON.parse(text)
  } catch {
    return { message: text }
  }
}

/**
 * 使用 /auth/refresh 刷新 accessToken（基于 swagger types.RefreshTokenRequest / LoginResponse）
 */
async function refreshAccessTokenOnce() {
  const refreshToken = getRefreshToken()
  if (!refreshToken) throw new Error("未找到 refreshToken，无法刷新登录状态")

  const r = await fetch(`${API_BASE}/auth/refresh`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify({ refresh_token: refreshToken })
  })
  const json = await parseJsonSafe(r)
  if (!r.ok) {
    const msg =
      json?.message ||
      json?.error ||
      `刷新登录状态失败（HTTP ${r.status}）`
    const err = new Error(msg)
    err.status = r.status
    err.body = json
    throw err
  }
  // swagger: types.LoginResponse { access_token, refresh_token, ... }
  const data = unwrapData(json)
  const access = data?.access_token || data?.accessToken
  const refresh = data?.refresh_token || data?.refreshToken
  if (!access) {
    throw new Error("刷新登录状态成功但响应中没有新的 access_token")
  }
  setTokens(access, refresh || refreshToken)
  return access
}

/**
 * @param {string} path 以 / 开头
 * @param {RequestInit & { skipAuth?: boolean, _retry?: boolean }} options
 */
export async function apiFetch(path, options = {}) {
  const { skipAuth, headers: optHeaders, _retry, ...rest } = options
  const url = path.startsWith("http") ? path : `${API_BASE}${path}`
  const headers = skipAuth
    ? { ...optHeaders }
    : { ...authHeaders(), ...optHeaders }

  const r = await fetch(url, { ...rest, headers })
  const json = await parseJsonSafe(r)

  // 如果 accessToken 失效（401）且未重试过，尝试使用 refresh_token 刷新一次
  if (!r.ok && r.status === 401 && !skipAuth && !_retry) {
    try {
      await refreshAccessTokenOnce()
    } catch (e) {
      // 刷新失败，直接抛出原 401 错误
      const msg =
        json?.message ||
        json?.error ||
        `请求失败（HTTP ${r.status}）`
      const err = new Error(msg)
      err.status = r.status
      err.body = json
      throw err
    }

    // 刷新成功后，带新 token 重试一次
    const retryHeaders = skipAuth
      ? { ...optHeaders }
      : { ...authHeaders(), ...optHeaders }
    const r2 = await fetch(url, { ...rest, headers: retryHeaders })
    const json2 = await parseJsonSafe(r2)
    if (!r2.ok) {
      const msg2 =
        json2?.message ||
        json2?.error ||
        `请求失败（HTTP ${r2.status}）`
      const err2 = new Error(msg2)
      err2.status = r2.status
      err2.body = json2
      throw err2
    }
    return json2
  }

  if (!r.ok) {
    const msg =
      json?.message ||
      json?.error ||
      `请求失败（HTTP ${r.status}）`
    const err = new Error(msg)
    err.status = r.status
    err.body = json
    throw err
  }
  return json
}
