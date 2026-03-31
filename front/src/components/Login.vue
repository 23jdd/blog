<template>
  <div class="login-wrap">
    <form class="card" @submit.prevent="handleSubmit">
      <h2 class="title">个人博客</h2>

      <div class="mode-switch">
        <button
          type="button"
          class="switch-btn"
          :class="{ active: !isRegister }"
          @click="isRegister = false"
        >
          登录
        </button>
        <button
          type="button"
          class="switch-btn"
          :class="{ active: isRegister }"
          @click="isRegister = true"
        >
          注册
        </button>
      </div>

      <div
        v-show="display"
        class="messagebox"
        :class="infoType"
        role="status"
        aria-live="polite"
      >
        {{ info }}
      </div>

      <div class="field">
        <label class="label" for="email">邮箱</label>
        <input
          id="email"
          type="email"
          autocomplete="email"
          placeholder="请输入邮箱"
          v-model="email"
        />
      </div>

      <div class="field">
        <label class="label" for="password">密码</label>
        <input
          id="password"
          type="password"
          minlength="6"
          maxlength="20"
          autocomplete="current-password"
          placeholder="请输入密码"
          v-model="password"
        />
      </div>

      <div class="field" v-if="isRegister">
        <label class="label" for="code">验证码</label>
        <input
          id="code"
          type="text"
          inputmode="numeric"
          placeholder="请输入验证码"
          v-model.trim="code"
        />
      </div>
      <div class="actions">
        <button type="submit" class="btn primary" :disabled="isSubmitting">
          {{ isSubmitting ? "提交中..." : isRegister ? "注册" : "登录" }}
        </button>
        <button
          v-if="isRegister"
          type="button"
          class="btn secondary"
          :disabled="!canSendCode"
          @click="SendCode"
        >
          {{ cooldown > 0 ? `${cooldown}s 后重发` : "发送验证码" }}
        </button>
      </div>
    </form>
  </div>
</template>
<script>
import { API_BASE } from "@/api/config.js"
const Baseurl = API_BASE
export default {
   data() {
      return {
            email:"",
            password:"",
            code:"",
            info:"",
            display:false,
            // messagebox 类型：error / success / info
            infoType:"info",
            isRegister: false,
            isSubmitting: false,
            cooldown: 0,
            cooldownTimer: null
      }
   },
   computed: {
     canSendCode() {
       return !this.isSubmitting && this.cooldown === 0
     }
   },
   async mounted() {
     const token = window.localStorage.getItem("accessToken")
     if (token) {
       await fetch(Baseurl + "/auth/judgeToken", {
         method: "GET",
         headers: {
           Authorization: `Bearer ${token}`
         }
       })
     }
   },
   beforeUnmount() {
     if (this.cooldownTimer) {
       clearInterval(this.cooldownTimer)
       this.cooldownTimer = null
     }
   },
   methods:{
     async Judgetoken(){
      let token = window.localStorage.getItem("accessToken")
      if (token) {
        let r = await fetch(Baseurl + "/auth/judgeToken", {
          method: "GET",
          headers: {
            Authorization: `Bearer ${token}`,
          }
        })
        let data = await r.json().catch(() => ({}))
        return data.code===200
      }
     },
      setMessage(message, type = "info") {
        this.info = message
        this.infoType = type
        this.display = true
        setTimeout(() => {
          this.display = false
        }, 3000) //3秒后自动消失
      },
      validateForm(state) {
        if (!this.email) return "请填写邮箱"
        if (!this.password) return "请填写密码"
        if (this.password.length < 6) return "密码长度至少为 6 位"
        if (state === "register" && !this.code) return "注册时请填写验证码"
        return ""
      },
      async handleSubmit() {
        const state = this.isRegister ? "register" : "login"
        const validationError = this.validateForm(state)
        if (validationError) {
          this.setMessage(validationError, "error")
          return
        }
        await this.Login(state)
      },
      async Login(state){
           let user={
                username:this.email,
                password:this.password,
           }
           if(state==="register"){
                user.code=this.code
           }
           console.log("call")
           this.display = false
           this.info = ""
           this.infoType = "info"
           this.isSubmitting = true
           try {
             let r = await fetch(Baseurl + `/auth/${state}`, {
               method: "POST",
               headers: {
                 "Content-Type": "application/json"
               },
               body: JSON.stringify(user),
             })
             let data=await r.json().catch(() => ({}))
             if (!r.ok) {
               throw new Error(data?.message || `${state === "register" ? "注册" : "登录"}失败（HTTP ${r.status}）`)
             }
             if (data?.access_token) {
               window.localStorage.setItem("accessToken", data.access_token)
             }
             if(data?.refresh_token) {
               window.localStorage.setItem("refreshToken", data.refresh_token)
             }
             console.log(data)
             this.setMessage(data?.message || `${state === "register" ? "注册" : "登录"}成功`, "success")
           }
           catch(e){
              this.setMessage(e?.message || `${state === "register" ? "注册" : "登录"}失败`, "error")
           } finally {
             this.isSubmitting = false
          
            if(await this.Judgetoken()) {
                window.location.href = "/main"      
            }
             console.log("judgetoken")
           }
      },
      startCooldown(seconds = 60) {
        if (this.cooldownTimer) {
          clearInterval(this.cooldownTimer)
          this.cooldownTimer = null
        }
        this.cooldown = seconds
        this.cooldownTimer = setInterval(() => {
          this.cooldown -= 1
          if (this.cooldown <= 0) {
            clearInterval(this.cooldownTimer)
            this.cooldownTimer = null
            this.cooldown = 0
          }
        }, 1000)
      },
      async SendCode(){
         try{
          if (!this.canSendCode) return
          this.display = false
          this.info = ""
          this.infoType = "info"
          if (!this.email) {
            this.setMessage("请先填写邮箱", "error")
            return
          }
          let r = await fetch(
            `${Baseurl}/verification/send?email=${encodeURIComponent(this.email)}`,
            { method: "GET" }
          )
           let data = await r.json().catch(() => ({}))
           if (!r.ok) {
            throw new Error(data?.message || `发送验证码失败（HTTP ${r.status}）`)
           }
           this.setMessage(data?.message || "验证码已发送，请注意查收", "success")
           this.startCooldown(60)
         }catch(e){
          this.setMessage(e?.message || "发送验证码失败", "error")
         }
      }
   }
 }
</script>
<style scoped>
.login-wrap {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background:
    radial-gradient(circle at 15% 10%, rgba(59, 130, 246, 0.18) 0%, transparent 35%),
    radial-gradient(circle at 85% 85%, rgba(16, 185, 129, 0.12) 0%, transparent 30%),
    linear-gradient(180deg, #020617 0%, #0f172a 55%, #111827 100%);
  padding: 24px;
}

.card {
  width: min(420px, 92vw);
  background: rgba(15, 23, 42, 0.86);
  border: 1px solid rgba(148, 163, 184, 0.24);
  border-radius: 18px;
  box-shadow:
    0 22px 50px rgba(2, 6, 23, 0.55),
    inset 0 1px 0 rgba(255, 255, 255, 0.04);
  padding: 30px 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.card,
.card * {
  box-sizing: border-box;
}

.title {
  margin: 0 0 2px;
  font-size: 20px;
  font-weight: 700;
  color: #f8fafc;
  text-align: center;
}

.mode-switch {
  display: grid;
  grid-template-columns: 1fr 1fr;
  background: rgba(30, 41, 59, 0.9);
  border-radius: 10px;
  padding: 4px;
  gap: 4px;
}

.switch-btn {
  border: 1px solid transparent;
  border-radius: 8px;
  background: transparent;
  padding: 8px 10px;
  font-weight: 600;
  cursor: pointer;
  color: #94a3b8;
}

.switch-btn.active {
  background: linear-gradient(180deg, rgba(37, 99, 235, 0.4) 0%, rgba(30, 64, 175, 0.4) 100%);
  border-color: rgba(96, 165, 250, 0.35);
  color: #dbeafe;
  box-shadow: 0 4px 12px rgba(30, 64, 175, 0.3);
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.label {
  font-size: 14px;
  color: #cbd5e1;
}

.card input {
  width: 100%;
  padding: 10px 12px;
  font-size: 14px;
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.28);
  background: rgba(2, 6, 23, 0.5);
  color: #f1f5f9;
  outline: none;
  transition: border-color 0.18s ease, box-shadow 0.18s ease;
}

.card input::placeholder {
  color: #64748b;
}

.card input:focus {
  border-color: #2563eb;
  box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.14);
}

.actions {
  display: flex;
  gap: 12px;
  margin-top: 4px;
  align-items: stretch;
}

.btn {
  flex: 1;
  min-height: 40px;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid transparent;
  cursor: pointer;
  font-weight: 600;
  transition: transform 0.08s ease, background-color 0.18s ease, border-color 0.18s ease, box-shadow 0.18s ease;
}

.btn:disabled {
  opacity: 0.62;
  cursor: not-allowed;
}

.btn:active {
  transform: translateY(1px);
}

.btn.primary {
  background: linear-gradient(135deg, #2563eb 0%, #3b82f6 100%);
  color: #ffffff;
  box-shadow: 0 8px 18px rgba(37, 99, 235, 0.28);
}

.btn.primary:hover {
  filter: brightness(1.05);
}

.btn.secondary {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(148, 163, 184, 0.32);
  color: #e2e8f0;
}

.btn.secondary:hover {
  border-color: rgba(148, 163, 184, 0.5);
  background: rgba(51, 65, 85, 0.88);
}

@media (max-width: 420px) {
  .actions {
    flex-direction: column;
  }
}

.messagebox{
  width: 100%;
  margin: 2px 0 0;
  padding: 9px 11px;
  border-radius: 10px;
  font-size: 13px;
  line-height: 1.35;
  border: 1px solid transparent;
  box-shadow: 0 6px 14px rgba(15, 23, 42, 0.08);
}

.messagebox.info{
  background: rgba(37, 99, 235, 0.18);
  border-color: rgba(96, 165, 250, 0.35);
  color: #bfdbfe;
}

.messagebox.success{
  background: rgba(16, 185, 129, 0.16);
  border-color: rgba(52, 211, 153, 0.35);
  color: #a7f3d0;
}

.messagebox.error{
  background: rgba(239, 68, 68, 0.16);
  border-color: rgba(248, 113, 113, 0.35);
  color: #fecaca;
}
</style>