<script setup>
import { ref, watch } from "vue"

const props = defineProps({
  title: {
    type: String,
    default: ""
  },
  activeKey: {
    type: String,
    default: "home"
  },
  items: {
    type: Array,
    default: () => [
      { key: "home", label: "Home" },
      { key: "page", label: "Page" },

    ]
  }
})

const emit = defineEmits(["select"])
const localActiveKey = ref(props.activeKey)

watch(
  () => props.activeKey,
  (next) => {
    localActiveKey.value = next
  }
)

function onSelect(it) {
  localActiveKey.value = it.key
  emit("select", it)
}
</script>

<template>
  <aside class="sidebar">
    <div class="brand">
    
      <div class="brand-text">
        <div class="brand-title">{{ props.title }}</div>
        <div class="brand-sub">Menu</div>
      </div>
    </div>

    <nav class="nav" aria-label="侧边导航">
      <button
        v-for="it in props.items"
        :key="it.key"
        type="button"
        class="nav-item"
        :class="{ active: it.key === localActiveKey }"
        @click="onSelect(it)"
      >
        <span class="dot" aria-hidden="true"></span>
        <span class="label">{{ it.label }}</span>
      </button>
    </nav>

    <div class="foot">
      <div class="muted">© 2018 - 2026</div>
      <div class="muted">All rights reserved.</div>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  width: 260px;
  height: 100vh;
  position: sticky;
  top: 0;
  padding: 18px 14px;
  background: linear-gradient(180deg, #111315 0%, #1a1d21 55%, #111315 100%);
  border-right: 1px solid rgba(163, 163, 163, 0.16);
  box-shadow: 14px 0 40px rgba(2, 6, 23, 0.35);
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 10px 14px;
  border-radius: 14px;
  background: rgba(38, 38, 38, 0.45);
  border: 1px solid rgba(163, 163, 163, 0.14);
}

.logo {
  width: 42px;
  height: 42px;
  border-radius: 12px;
  background: radial-gradient(circle at 30% 30%, rgba(212, 212, 212, 0.9), rgba(115, 115, 115, 0.15));
  border: 1px solid rgba(163, 163, 163, 0.35);
  box-shadow: 0 10px 20px rgba(2, 6, 23, 0.35), inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.brand-text {
  display: grid;
  gap: 2px;
}

.brand-title {
  font-weight: 900;
  letter-spacing: 0.2px;
  color: #e5e7eb;
  line-height: 1.1;
}

.brand-sub {
  font-size: 12px;
  color: #a3a3a3;
}

.nav {
  display: grid;
  gap: 6px;
  padding: 6px 2px;
}

.nav-item {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 12px;
  cursor: pointer;
  border: 1px solid transparent;
  background: transparent;
  color: #cbd5e1;
  text-align: left;
  transition: background-color 0.16s ease, border-color 0.16s ease, transform 0.16s ease;
}

.nav-item:hover {
  background: rgba(51, 51, 51, 0.55);
  border-color: rgba(163, 163, 163, 0.16);
  transform: translateX(2px);
}

.nav-item.active {
  background: rgba(82, 82, 82, 0.28);
  border-color: rgba(163, 163, 163, 0.28);
  color: #f5f5f5;
}

.dot {
  width: 10px;
  height: 10px;
  border-radius: 999px;
  background: rgba(163, 163, 163, 0.45);
  box-shadow: 0 0 0 4px rgba(163, 163, 163, 0.08);
}

.nav-item.active .dot {
  background: #d4d4d4;
  box-shadow: 0 0 0 4px rgba(212, 212, 212, 0.18);
}

.label {
  font-weight: 700;
  letter-spacing: 0.2px;
}

.foot {
  margin-top: auto;
  padding: 12px 10px 6px;
  border-top: 1px solid rgba(163, 163, 163, 0.14);
  display: grid;
  gap: 2px;
}

.muted {
  font-size: 12px;
  color: #737373;
}
</style>
