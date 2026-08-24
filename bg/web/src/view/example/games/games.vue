
<template>
  <div>
    <div class="gva-search-box">
      <el-form ref="elSearchFormRef" :inline="true" :model="searchInfo" class="demo-form-inline" @keyup.enter="onSubmit">
            <el-form-item label="Platform" prop="platformCode">
  <el-input v-model="searchInfo.platformCode" placeholder="HEDOC / M7 / M7PP" />
</el-form-item>
            <el-form-item label="关联厂商ID (game_providers.id)" prop="providerId">
  <el-input v-model.number="searchInfo.providerId" placeholder="搜索条件" />
</el-form-item>
            
            <el-form-item label="关联分类ID (game_types.id)" prop="gameTypeId">
  <el-input v-model.number="searchInfo.gameTypeId" placeholder="搜索条件" />
</el-form-item>
            

        <template v-if="showAllQuery">
          <!-- 将需要控制显示状态的查询条件添加到此范围内 -->
        </template>

        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit">查询</el-button>
          <el-button icon="refresh" @click="onReset">重置</el-button>
          <el-button link type="primary" icon="arrow-down" @click="showAllQuery=true" v-if="!showAllQuery">展开</el-button>
          <el-button link type="primary" icon="arrow-up" @click="showAllQuery=false" v-else>收起</el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="gva-table-box">
        <div class="gva-btn-list">
            <el-button  type="primary" icon="plus" @click="openDialog()">新增</el-button>
            <el-button  icon="delete" style="margin-left: 10px;" :disabled="!multipleSelection.length" @click="onDelete">删除</el-button>
            
        </div>
        <el-table
        ref="multipleTable"
        style="width: 100%"
        tooltip-effect="dark"
        :data="tableData"
        row-key="id"
        @selection-change="handleSelectionChange"
        >
        <el-table-column type="selection" width="55" />
        
            <el-table-column align="left" label="id字段" prop="id" width="120" />

            <el-table-column align="left" label="Platform" prop="platformCode" width="120" />

            <el-table-column align="left" label="关联厂商ID (game_providers.id)" prop="providerId" width="120" />

            <el-table-column align="left" label="关联分类ID (game_types.id)" prop="gameTypeId" width="120" />

            <el-table-column align="left" label="厂商侧的游戏ID" prop="gameCode" width="120" />

            <el-table-column label="游戏名称(多语言)" prop="name" width="200">
    <template #default="scope">
        [JSON]
    </template>
</el-table-column>
            <el-table-column align="left" label="封面图片URL" prop="imgUrl" width="120" />

            <el-table-column align="left" label="点击/热度" prop="views" width="120" />

            <el-table-column align="left" label="排序, 越大越前" prop="sort" width="120" />

            <el-table-column align="left" label="1:上架 0:下架" prop="status" width="120">
    <template #default="scope">{{ formatBoolean(scope.row.status) }}</template>
</el-table-column>
            <el-table-column align="left" label="createdAt字段" prop="createdAt" width="180">
   <template #default="scope">{{ formatDate(scope.row.createdAt) }}</template>
</el-table-column>
            <el-table-column align="left" label="updatedAt字段" prop="updatedAt" width="180">
   <template #default="scope">{{ formatDate(scope.row.updatedAt) }}</template>
</el-table-column>
        <el-table-column align="left" label="操作" fixed="right" :min-width="appStore.operateMinWith">
            <template #default="scope">
            <el-button  type="primary" link class="table-button" @click="getDetails(scope.row)"><el-icon style="margin-right: 5px"><InfoFilled /></el-icon>查看</el-button>
            <el-button  type="primary" link icon="edit" class="table-button" @click="updateGamesFunc(scope.row)">编辑</el-button>
            <el-button   type="primary" link icon="delete" @click="deleteRow(scope.row)">删除</el-button>
            </template>
        </el-table-column>
        </el-table>
        <div class="gva-pagination">
            <el-pagination
            layout="total, sizes, prev, pager, next, jumper"
            :current-page="page"
            :page-size="pageSize"
            :page-sizes="[10, 30, 50, 100]"
            :total="total"
            @current-change="handleCurrentChange"
            @size-change="handleSizeChange"
            />
        </div>
    </div>
    <el-drawer destroy-on-close :size="appStore.drawerSize" v-model="dialogFormVisible" :show-close="false" :before-close="closeDialog">
       <template #header>
              <div class="flex justify-between items-center">
                <span class="text-lg">{{type==='create'?'新增':'编辑'}}</span>
                <div>
                  <el-button :loading="btnLoading" type="primary" @click="enterDialog">确 定</el-button>
                  <el-button @click="closeDialog">取 消</el-button>
                </div>
              </div>
            </template>

          <el-form :model="formData" label-position="top" ref="elFormRef" :rules="rule" label-width="80px">
            <el-form-item label="Platform:" prop="platformCode">
    <el-input v-model="formData.platformCode" :clearable="true" placeholder="HEDOC / M7 / M7PP" />
</el-form-item>
            <el-form-item label="关联厂商ID (game_providers.id):" prop="providerId">
    <el-input v-model.number="formData.providerId" :clearable="true" placeholder="请输入关联厂商ID (game_providers.id)" />
</el-form-item>
            <el-form-item label="关联分类ID (game_types.id):" prop="gameTypeId">
    <el-input v-model.number="formData.gameTypeId" :clearable="true" placeholder="请输入关联分类ID (game_types.id)" />
</el-form-item>
            <el-form-item label="厂商侧的游戏ID:" prop="gameCode">
    <el-input v-model="formData.gameCode" :clearable="true" placeholder="请输入厂商侧的游戏ID" />
</el-form-item>
            <el-form-item label="游戏名称(多语言):" prop="name">
    // 此字段为json结构，可以前端自行控制展示和数据绑定模式 需绑定json的key为 formData.name 后端会按照json的类型进行存取
    {{ formData.name }}
</el-form-item>
            <el-form-item label="封面图片URL:" prop="imgUrl">
    <el-input v-model="formData.imgUrl" :clearable="true" placeholder="请输入封面图片URL" />
</el-form-item>
            <el-form-item label="点击/热度:" prop="views">
    <el-input v-model.number="formData.views" :clearable="true" placeholder="请输入点击/热度" />
</el-form-item>
            <el-form-item label="排序, 越大越前:" prop="sort">
    <el-input v-model.number="formData.sort" :clearable="true" placeholder="请输入排序, 越大越前" />
</el-form-item>
            <el-form-item label="1:上架 0:下架:" prop="status">
    <el-switch v-model="formData.status" active-color="#13ce66" inactive-color="#ff4949" active-text="是" inactive-text="否" clearable ></el-switch>
</el-form-item>
            <el-form-item label="createdAt字段:" prop="createdAt">
    <el-date-picker v-model="formData.createdAt" type="date" style="width:100%" placeholder="选择日期" :clearable="true" />
</el-form-item>
            <el-form-item label="updatedAt字段:" prop="updatedAt">
    <el-date-picker v-model="formData.updatedAt" type="date" style="width:100%" placeholder="选择日期" :clearable="true" />
</el-form-item>
          </el-form>
    </el-drawer>

    <el-drawer destroy-on-close :size="appStore.drawerSize" v-model="detailShow" :show-close="true" :before-close="closeDetailShow" title="查看">
            <el-descriptions :column="1" border>
                    <el-descriptions-item label="id字段">
    {{ detailFrom.id }}
</el-descriptions-item>
                    <el-descriptions-item label="Platform">
    {{ detailFrom.platformCode }}
</el-descriptions-item>
                    <el-descriptions-item label="关联厂商ID (game_providers.id)">
    {{ detailFrom.providerId }}
</el-descriptions-item>
                    <el-descriptions-item label="关联分类ID (game_types.id)">
    {{ detailFrom.gameTypeId }}
</el-descriptions-item>
                    <el-descriptions-item label="厂商侧的游戏ID">
    {{ detailFrom.gameCode }}
</el-descriptions-item>
                    <el-descriptions-item label="游戏名称(多语言)">
    {{ detailFrom.name }}
</el-descriptions-item>
                    <el-descriptions-item label="封面图片URL">
    {{ detailFrom.imgUrl }}
</el-descriptions-item>
                    <el-descriptions-item label="点击/热度">
    {{ detailFrom.views }}
</el-descriptions-item>
                    <el-descriptions-item label="排序, 越大越前">
    {{ detailFrom.sort }}
</el-descriptions-item>
                    <el-descriptions-item label="1:上架 0:下架">
    {{ detailFrom.status }}
</el-descriptions-item>
                    <el-descriptions-item label="createdAt字段">
    {{ detailFrom.createdAt }}
</el-descriptions-item>
                    <el-descriptions-item label="updatedAt字段">
    {{ detailFrom.updatedAt }}
</el-descriptions-item>
            </el-descriptions>
        </el-drawer>

  </div>
</template>

<script setup>
import {
  createGames,
  deleteGames,
  deleteGamesByIds,
  updateGames,
  findGames,
  getGamesList
} from '@/api/example/games'

// 全量引入格式化工具 请按需保留
import { getDictFunc, formatDate, formatBoolean, filterDict ,filterDataSource, returnArrImg, onDownloadFile } from '@/utils/format'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ref, reactive } from 'vue'
import { useAppStore } from "@/pinia"




defineOptions({
    name: 'Games'
})

// 提交按钮loading
const btnLoading = ref(false)
const appStore = useAppStore()

// 控制更多查询条件显示/隐藏状态
const showAllQuery = ref(false)

// 自动化生成的字典（可能为空）以及字段
const formData = ref({
            platformCode: 'HEDOC',
            providerId: undefined,
            gameTypeId: undefined,
            gameCode: '',
            name: {},
            imgUrl: '',
            views: undefined,
            sort: undefined,
            status: false,
            createdAt: new Date(),
            updatedAt: new Date(),
        })



// 验证规则
const rule = reactive({
})

const elFormRef = ref()
const elSearchFormRef = ref()

// =========== 表格控制部分 ===========
const page = ref(1)
const total = ref(0)
const pageSize = ref(10)
const tableData = ref([])
const searchInfo = ref({})
// 重置
const onReset = () => {
  searchInfo.value = {}
  getTableData()
}

// 搜索
const onSubmit = () => {
  elSearchFormRef.value?.validate(async(valid) => {
    if (!valid) return
    page.value = 1
    if (searchInfo.value.status === ""){
        searchInfo.value.status=null
    }
    getTableData()
  })
}

// 分页
const handleSizeChange = (val) => {
  pageSize.value = val
  getTableData()
}

// 修改页面容量
const handleCurrentChange = (val) => {
  page.value = val
  getTableData()
}

// 查询
const getTableData = async() => {
  const table = await getGamesList({ page: page.value, pageSize: pageSize.value, ...searchInfo.value })
  if (table.code === 0) {
    tableData.value = table.data.list
    total.value = table.data.total
    page.value = table.data.page
    pageSize.value = table.data.pageSize
  }
}

getTableData()

// ============== 表格控制部分结束 ===============

// 获取需要的字典 可能为空 按需保留
const setOptions = async () =>{
}

// 获取需要的字典 可能为空 按需保留
setOptions()


// 多选数据
const multipleSelection = ref([])
// 多选
const handleSelectionChange = (val) => {
    multipleSelection.value = val
}

// 删除行
const deleteRow = (row) => {
    ElMessageBox.confirm('确定要删除吗?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
            deleteGamesFunc(row)
        })
    }

// 多选删除
const onDelete = async() => {
  ElMessageBox.confirm('确定要删除吗?', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async() => {
      const ids = []
      if (multipleSelection.value.length === 0) {
        ElMessage({
          type: 'warning',
          message: '请选择要删除的数据'
        })
        return
      }
      multipleSelection.value &&
        multipleSelection.value.map(item => {
          ids.push(item.id)
        })
      const res = await deleteGamesByIds({ ids })
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: '删除成功'
        })
        if (tableData.value.length === ids.length && page.value > 1) {
          page.value--
        }
        getTableData()
      }
      })
    }

// 行为控制标记（弹窗内部需要增还是改）
const type = ref('')

// 更新行
const updateGamesFunc = async(row) => {
    const res = await findGames({ id: row.id })
    type.value = 'update'
    if (res.code === 0) {
        formData.value = res.data
        dialogFormVisible.value = true
    }
}


// 删除行
const deleteGamesFunc = async (row) => {
    const res = await deleteGames({ id: row.id })
    if (res.code === 0) {
        ElMessage({
                type: 'success',
                message: '删除成功'
            })
            if (tableData.value.length === 1 && page.value > 1) {
            page.value--
        }
        getTableData()
    }
}

// 弹窗控制标记
const dialogFormVisible = ref(false)

// 打开弹窗
const openDialog = () => {
    type.value = 'create'
    dialogFormVisible.value = true
}

// 关闭弹窗
const closeDialog = () => {
    dialogFormVisible.value = false
    formData.value = {
        platformCode: 'HEDOC',
        providerId: undefined,
        gameTypeId: undefined,
        gameCode: '',
        name: {},
        imgUrl: '',
        views: undefined,
        sort: undefined,
        status: false,
        createdAt: new Date(),
        updatedAt: new Date(),
        }
}
// 弹窗确定
const enterDialog = async () => {
     btnLoading.value = true
     elFormRef.value?.validate( async (valid) => {
             if (!valid) return btnLoading.value = false
              let res
              switch (type.value) {
                case 'create':
                  res = await createGames(formData.value)
                  break
                case 'update':
                  res = await updateGames(formData.value)
                  break
                default:
                  res = await createGames(formData.value)
                  break
              }
              btnLoading.value = false
              if (res.code === 0) {
                ElMessage({
                  type: 'success',
                  message: '创建/更改成功'
                })
                closeDialog()
                getTableData()
              }
      })
}

const detailFrom = ref({})

// 查看详情控制标记
const detailShow = ref(false)


// 打开详情弹窗
const openDetailShow = () => {
  detailShow.value = true
}


// 打开详情
const getDetails = async (row) => {
  // 打开弹窗
  const res = await findGames({ id: row.id })
  if (res.code === 0) {
    detailFrom.value = res.data
    openDetailShow()
  }
}


// 关闭详情弹窗
const closeDetailShow = () => {
  detailShow.value = false
  detailFrom.value = {}
}


</script>

<style>

</style>
