<template>
  <div class="domain-stats" v-loading="loading">
    <div class="toolbar">
      <el-select v-model="domain" clearable placeholder="选择域名" style="width:220px"><el-option v-for="item in domains" :key="item" :label="item" :value="item" /></el-select>
      <el-date-picker v-model="range" type="daterange" value-format="YYYY-MM-DD" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" :clearable="false" />
      <el-button type="primary" :icon="Search" @click="load">查询</el-button>
    </div>
    <el-table :data="rows" stripe border>
      <el-table-column prop="domain" label="域名" min-width="220" />
      <el-table-column prop="date" label="日期" width="150"><template #default="{ row }">{{ String(row.date || '').slice(0, 10) }}</template></el-table-column>
      <el-table-column prop="clicks" label="点击次数" width="160" align="right" />
      <el-table-column prop="registrations" label="注册人数" width="160" align="right" />
    </el-table>
    <el-empty v-if="!loading && !rows.length" description="暂无统计数据" />
  </div>
</template>
<script setup>
import { computed, onMounted, ref } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getDomainStats } from '@/api/example/domainStats'
defineOptions({ name: 'DomainStats' })
const loading = ref(false), domain = ref(''), range = ref([]), rows = ref([]), domains = ref([])
const format = date => { const d = new Date(date); return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}` }
const load = async () => { loading.value = true; try { const params = { from: range.value[0], to: range.value[1] }; if (domain.value) params.domain = domain.value; const res = await getDomainStats(params); const data = res.data || res; rows.value = data.items || []; domains.value = data.domains || domains.value } catch (e) { ElMessage.error(e?.message || '获取域名统计失败') } finally { loading.value = false } }
onMounted(() => { const today = format(new Date()); range.value = [today, today]; load() })
</script>
<style scoped>.domain-stats{padding:16px}.toolbar{display:flex;gap:10px;flex-wrap:wrap;margin-bottom:16px;background:#fff;padding:16px;border:1px solid #edf0f5;border-radius:8px}</style>
