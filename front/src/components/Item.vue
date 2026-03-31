<script setup>
import { computed } from "vue"

const props = defineProps({
  item: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(["click"])

const hasAvatar = computed(() => {
  const u = props.item?.avatar
  return typeof u === "string" && u.trim().length > 0
})

const fallbackLetter = computed(() => {
  const name = props.item?.name || ""
  const ch = name.trim().charAt(0)
  return ch || "·"
})
</script>

<template>
  <article class="friend-card" @click="emit('click', props.item)">
    <aside class="avatar-aside" aria-hidden="true">
      <img
        v-if="hasAvatar"
        class="avatar"
        :src="props.item.avatar"
        :alt="props.item.name || '封面'"
      />
      <span v-else class="avatar-fallback">{{ fallbackLetter }}</span>
    </aside>

    <div class="main">
      <div class="top">
        <h4 class="name">{{ props.item.name }}</h4>
        <span v-if="props.item.tag" class="tag">{{ props.item.tag }}</span>
      </div>
      <p v-if="props.item.desc" class="desc">{{ props.item.desc }}</p>
    </div>
  </article>
</template>

<style scoped>
.friend-card {
  display: grid;
  grid-template-columns: 64px 1fr;
  gap: 12px;
  align-items: center;
  padding: 12px 14px;
  border-radius: 12px;
  background: rgba(15, 23, 42, 0.55);
  border: 1px solid rgba(148, 163, 184, 0.16);
  box-shadow: 0 12px 26px rgba(2, 6, 23, 0.35);
  cursor: pointer;
  transition: transform 0.16s ease, border-color 0.16s ease, background-color 0.16s ease;
}

.friend-card:hover {
  transform: translateY(-2px);
  border-color: rgba(96, 165, 250, 0.35);
  background: rgba(15, 23, 42, 0.7);
}

.friend-card:active {
  transform: translateY(-1px);
}

/* aside：左侧头像图片区块效果 */
.avatar-aside {
  width: 64px;
  height: 64px;
  border-radius: 14px;
  background: linear-gradient(180deg, rgba(30, 41, 59, 0.9), rgba(15, 23, 42, 0.6));
  border: 1px solid rgba(148, 163, 184, 0.18);
  display: grid;
  place-items: center;
  overflow: hidden;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.06),
    0 10px 18px rgba(2, 6, 23, 0.35);
}

.avatar {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  object-fit: cover;
  filter: saturate(1.05) contrast(1.02);
}

.avatar-fallback {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  font-size: 20px;
  font-weight: 800;
  color: #94a3b8;
  background: rgba(30, 41, 59, 0.85);
  border: 1px solid rgba(148, 163, 184, 0.2);
}

.main {
  min-width: 0;
  display: grid;
  gap: 6px;
}

.top {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 10px;
  min-width: 0;
}

.name {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
  color: #e5e7eb;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tag {
  flex: none;
  font-size: 12px;
  color: #93c5fd;
  background: rgba(37, 99, 235, 0.12);
  border: 1px solid rgba(96, 165, 250, 0.25);
  padding: 2px 8px;
  border-radius: 999px;
}

.desc {
  margin: 0;
  font-size: 13px;
  color: #9ca3af;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
