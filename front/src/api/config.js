/** 与 swagger host 一致，可通过环境变量覆盖 */
export const API_BASE =
  (typeof import.meta !== "undefined" && import.meta.env?.VITE_API_BASE) ||
  "http://localhost:8080"
