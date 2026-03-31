<script>
export default {
    name: "UserInfo",
    emits: ["update:user", "save:personImage"],
    props: {
        user: {
            type: Object,
            default: () => ({
                avatar: "https://th.bing.com/th/id/OIP.gDtT0kmCsTfTeQ2wv-JTiQAAAA?w=208&h=208&c=7&r=0&o=7&dpr=1.5&pid=1.7&rm=3",
                nickname: "long",
                email: "1234567890@qq.com",
                age: null,
                gender: "",
                bio: "这个人很神秘，还没有留下简介。",
                createdAt: ""
            })
        }
    },
    data() {
        return {
            // 使用本地副本，避免直接修改 props
            localUser: { ...this.user },
            isEditingProfile: false,
            formAge: "",
            formGender: ""
        }
    },
    watch: {
        user: {
            handler(nextUser) {
                this.localUser = { ...nextUser }
            },
            deep: true
        }
    },
    computed: {
        displayName() {
            return this.localUser?.nickname || "未命名用户"
        },
        displayEmail() {
            return this.localUser?.email || "暂无邮箱"
        },
        displayAge() {
            const age = this.localUser?.age
            if (age === null || age === undefined || age === "") return "未知"
            return `${age} 岁`
        },
        displayGender() {
            const gender = this.localUser?.gender
            if (!gender) return "未填写"
            const map = {
                male: "男",
                female: "女",
                other: "其他"
            }
            return map[gender] || gender
        },
        displayBio() {
            return this.localUser?.bio || "这个人很神秘，还没有留下简介。"
        },
        formattedCreatedAt() {
            if (!this.localUser?.createdAt) return "未知"
            const date = new Date(this.localUser.createdAt)
            if (Number.isNaN(date.getTime())) return "未知"
            return date.toLocaleDateString("zh-CN")
        },
        avatarUrl() {
            return this.localUser?.avatar || "https://c-ssl.dtstatic.com/uploads/item/202004/14/20200414210224_dnzpo.thumb.400_0.jpg"
        }
    },
    methods: {
        chooseAvatarFile() {
            this.$refs.avatarInput?.click()
        },
        onAvatarFileChange(event) {
            const file = event.target?.files?.[0]
            if (!file) return
            if (!file.type.startsWith("image/")) {
                window.alert("请选择图片文件作为头像。")
                event.target.value = ""
                return
            }
            // 本地预览：使用浏览器生成的 blob URL 展示图片
            let formData = new FormData()
            formData.append("image", file)
            this.$emit("save:personImage", formData)
            event.target.value = ""
        },
        startEditAgeGender() {
            this.formAge = this.localUser?.age ?? ""
            this.formGender = this.localUser?.gender || ""
            this.isEditingProfile = true
        },
        cancelEditAgeGender() {
            this.isEditingProfile = false
            this.formAge = ""
            this.formGender = ""
        },
        saveAgeGender() {
            const parsedAge = Number(this.formAge)
            const validAge =
                Number.isInteger(parsedAge) && parsedAge >= 1 && parsedAge <= 150
                    ? parsedAge
                    : null
            this.localUser = {
                ...this.localUser,
                age: validAge,
                gender: this.formGender || ""
            }
            this.$emit("update:user", this.localUser)
            this.isEditingProfile = false
        }
    }
}
</script>

<template>
    <section class="user-card">
        <div class="header">
            <img :src="avatarUrl" alt="用户头像" class="avatar" />
            <div class="base">
                <h3 class="name">{{ displayName }}</h3>
                <p class="email">{{ displayEmail }}</p>
            </div>
        </div>

        <div class="content">
            <div class="info-grid">
                <div class="info-item">
                    <span class="info-label">年龄</span>
                    <span class="info-value">{{ displayAge }}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">性别</span>
                    <span class="info-value">{{ displayGender }}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">注册时间</span>
                    <span class="info-value">{{ formattedCreatedAt }}</span>
                </div>
            </div>
            <p class="bio">{{ displayBio }}</p>

            <div v-if="isEditingProfile" class="edit-panel">
                <label class="field">
                    <span class="field-label">年龄</span>
                    <input
                        v-model.number="formAge"
                        type="number"
                        min="1"
                        max="150"
                        placeholder="请输入年龄"
                    />
                </label>
                <label class="field">
                    <span class="field-label">性别</span>
                    <select v-model="formGender">
                        <option value>未填写</option>
                        <option value="male">男</option>
                        <option value="female">女</option>
                        <option value="other">其他</option>
                    </select>
                </label>
                <div class="edit-actions">
                    <button type="button" class="btn primary" @click="saveAgeGender">保存</button>
                    <button type="button" class="btn secondary" @click="cancelEditAgeGender">取消</button>
                </div>
            </div>

            <div class="actions">
                <input
                    ref="avatarInput"
                    type="file"
                    accept="image/*"
                    class="hidden-file"
                    @change="onAvatarFileChange"
                />
                <button type="button" class="btn primary" @click="chooseAvatarFile">修改头像</button>
                <button type="button" class="btn secondary" @click="startEditAgeGender">修改年龄和性别</button>
            </div>
        </div>
    </section>
</template>

<style scoped>
.user-card {
    width: min(760px, 96vw);
    min-height: 420px;
    margin: 24px auto;
    padding: 28px;
    border-radius: 18px;
    border: 1px solid rgba(148, 163, 184, 0.2);
    background: linear-gradient(160deg, #0f172a 0%, #111827 55%, #1e293b 100%);
    box-shadow: 0 20px 45px rgba(2, 6, 23, 0.45),
        inset 0 1px 0 rgba(255, 255, 255, 0.05);
}

.header {
    display: flex;
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
}

.avatar {
    width: 76px;
    height: 76px;
    border-radius: 50%;
    object-fit: cover;
    border: 2px solid rgba(96, 165, 250, 0.45);
    box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.15);
}

.base {
    display: flex;
    flex-direction: column;
    gap: 4px;
    width: 100%;
}

.name {
    margin: 0;
    font-size: 21px;
    color: #f8fafc;
    letter-spacing: 0.2px;
}

.email {
    margin: 0;
    font-size: 14px;
    color: #94a3b8;
}

.content {
    margin-top: 20px;
    display: flex;
    flex-direction: column;
    gap: 14px;
}

.info-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 10px;
}

.info-item {
    padding: 10px 12px;
    border-radius: 10px;
    background: rgba(15, 23, 42, 0.42);
    border: 1px solid rgba(148, 163, 184, 0.18);
    display: flex;
    flex-direction: column;
    gap: 6px;
}

.info-label {
    font-size: 12px;
    color: #93a6bf;
}

.info-value {
    font-size: 14px;
    font-weight: 600;
    color: #e2e8f0;
}

.bio {
    margin: 0;
    color: #cbd5e1;
    line-height: 1.8;
    padding: 10px 12px;
    border-radius: 10px;
    background: rgba(15, 23, 42, 0.35);
    border: 1px solid rgba(148, 163, 184, 0.14);
}

.meta {
    margin: 0;
    font-size: 13px;
    color: #9ca3af;
}

.actions {
    margin-top: 12px;
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.hidden-file {
    display: none;
}

.edit-panel {
    margin-top: 8px;
    padding: 12px;
    border-radius: 12px;
    background: rgba(15, 23, 42, 0.55);
    border: 1px solid rgba(148, 163, 184, 0.2);
    display: grid;
    gap: 10px;
}

.field {
    display: grid;
    gap: 6px;
}

.field-label {
    font-size: 13px;
    color: #cbd5e1;
}

.field input,
.field select {
    height: 34px;
    border: 1px solid rgba(148, 163, 184, 0.32);
    border-radius: 8px;
    padding: 0 10px;
    background: rgba(15, 23, 42, 0.7);
    color: #f8fafc;
    outline: none;
    transition: border-color 0.18s ease, box-shadow 0.18s ease;
}

.field input:focus,
.field select:focus {
    border-color: #60a5fa;
    box-shadow: 0 0 0 3px rgba(96, 165, 250, 0.2);
}

.edit-actions {
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.btn {
    border: 1px solid transparent;
    border-radius: 10px;
    padding: 8px 13px;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.16s ease;
    width: 100%;
}

.btn.primary {
    background: linear-gradient(135deg, #2563eb 0%, #3b82f6 100%);
    color: #ffffff;
}

.btn.primary:hover {
    filter: brightness(1.08);
    transform: translateY(-1px);
}

.btn.secondary {
    background: rgba(30, 41, 59, 0.68);
    border-color: rgba(148, 163, 184, 0.32);
    color: #e2e8f0;
}

.btn.secondary:hover {
    background: rgba(51, 65, 85, 0.85);
    border-color: rgba(148, 163, 184, 0.52);
}

@media (max-width: 480px) {
    .user-card {
        min-height: auto;
        padding: 20px;
    }
}
</style>