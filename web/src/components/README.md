# dashGO UI Components - Flat Design & Monochrome Theme

## 设计原则

本UI组件库遵循扁平化设计和黑白配色方案，提供简洁、专业的用户界面。

### 核心设计原则

1. **扁平化设计**
   - 最小圆角半径（2-4px）
   - 无渐变背景
   - 无发光效果
   - 简单的几何形状

2. **黑白配色**
   - 主背景：纯白色 (#FFFFFF)
   - 主文本：纯黑色 (#000000)
   - 层次区分：灰度色阶
   - 功能性颜色：仅用于状态指示

3. **简洁交互**
   - 简单的过渡效果（≤300ms）
   - 无复杂动画
   - 清晰的视觉反馈

## 组件使用

### Button 按钮

```vue
<Button variant="primary">主要按钮</Button>
<Button variant="secondary">次要按钮</Button>
<Button variant="outline">轮廓按钮</Button>
```

**变体 (variant)**
- `primary`: 黑色背景，白色文字
- `secondary`: 白色背景，黑色文字，灰色边框
- `outline`: 透明背景，黑色边框
- `ghost`: 透明背景，无边框

**尺寸 (size)**
- `sm`: 小尺寸
- `md`: 中等尺寸（默认）
- `lg`: 大尺寸

### Card 卡片

```vue
<Card>
  <template #header>标题</template>
  内容区域
  <template #footer>底部</template>
</Card>
```

**变体 (variant)**
- `default`: 标准卡片
- `bordered`: 加粗边框
- `elevated`: 带阴影

### Input 输入框

```vue
<Input
  v-model="value"
  label="标签"
  placeholder="请输入..."
  :error="errorMessage"
/>
```

### Table 表格

```vue
<Table :columns="columns" :data="data">
  <template #cell-action="{ row }">
    <Button size="sm">操作</Button>
  </template>
</Table>
```

### Badge 徽章

```vue
<Badge variant="success">成功</Badge>
<Badge variant="warning">警告</Badge>
<Badge variant="error">错误</Badge>
<Badge variant="info">信息</Badge>
```

## 设计令牌

所有设计令牌定义在 `src/styles/tokens.css` 中，包括：

- 颜色系统
- 间距比例
- 字体系统
- 圆角半径
- 阴影效果
- 过渡动画

## 主题管理

使用 `useTheme` 组合式函数管理主题：

```ts
import { useTheme } from '@/composables/useTheme'

const { theme, setTheme, toggleTheme } = useTheme()
```

## 响应式设计

所有组件都支持响应式设计，在移动端保持相同的扁平化风格和黑白配色。

断点定义：
- `sm`: 640px
- `md`: 768px
- `lg`: 1024px
- `xl`: 1280px
- `2xl`: 1536px

## 无障碍性

- 所有交互元素最小触摸目标：44px
- 键盘导航支持
- 屏幕阅读器友好
- 符合WCAG 2.1 AA标准的颜色对比度
