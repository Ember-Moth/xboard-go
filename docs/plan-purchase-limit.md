# 套餐购买数量限制设计

## 需求说明

套餐应该有购买数量限制功能，用于：
1. 限制套餐的最大销售数量（例如：限量100份）
2. 显示已售出数量和剩余数量
3. 当达到限制时，自动停止销售

## 当前问题

1. **Plan** 模型中有 `capacity_limit` 字段，但含义不明确
2. 没有记录已售出数量的字段
3. 流量配置已经在 Plan 中，但 User 也有重复字段

## 设计方案

### 1. Plan 模型字段说明

```go
type Plan struct {
    // ... 现有字段
    
    // 流量和限制配置（已存在）
    TransferEnable  int64  `json:"transfer_enable"`  // 流量配额（字节）
    SpeedLimit      *int   `json:"speed_limit"`      // 速度限制（Mbps）
    DeviceLimit     *int   `json:"device_limit"`     // 设备数量限制
    
    // 购买数量限制（需要明确）
    CapacityLimit   *int   `json:"capacity_limit"`   // 最大可售数量（null=不限制）
    SoldCount       int    `json:"sold_count"`       // 已售出数量（新增）
    
    // ... 其他字段
}
```

### 2. 字段含义

- **capacity_limit**: 最大可售数量
  - `null` 或 `0` = 不限制
  - `> 0` = 限制最大销售数量
  
- **sold_count**: 已售出数量（新增字段）
  - 每次购买套餐时 +1
  - 用户退订时 -1（可选）
  - 用于计算剩余数量

### 3. 购买逻辑

```go
// 检查是否可以购买
func (p *Plan) CanPurchase() bool {
    // 如果没有设置限制，可以购买
    if p.CapacityLimit == nil || *p.CapacityLimit <= 0 {
        return true
    }
    
    // 检查是否还有剩余
    return p.SoldCount < *p.CapacityLimit
}

// 获取剩余数量
func (p *Plan) GetRemainingCount() int {
    if p.CapacityLimit == nil || *p.CapacityLimit <= 0 {
        return -1 // 表示不限制
    }
    
    remaining := *p.CapacityLimit - p.SoldCount
    if remaining < 0 {
        return 0
    }
    return remaining
}
```

### 4. 数据库迁移

创建迁移文件 `005_add_plan_sold_count.sql`:

```sql
-- 添加已售出数量字段
ALTER TABLE `v2_plan` ADD COLUMN `sold_count` INT NOT NULL DEFAULT 0 COMMENT '已售出数量';

-- 初始化已售出数量（统计当前使用该套餐的用户数）
UPDATE `v2_plan` p 
SET `sold_count` = (
    SELECT COUNT(*) 
    FROM `v2_user` u 
    WHERE u.`plan_id` = p.`id`
);

-- 添加索引（可选，用于快速查询）
CREATE INDEX `idx_plan_capacity` ON `v2_plan`(`capacity_limit`, `sold_count`);
```

### 5. API 响应示例

**获取套餐列表**:
```json
{
  "data": [
    {
      "id": 1,
      "name": "基础套餐",
      "transfer_enable": 107374182400,  // 100GB
      "speed_limit": 100,                // 100Mbps
      "device_limit": 3,                 // 3台设备
      "capacity_limit": 100,             // 限量100份
      "sold_count": 85,                  // 已售85份
      "remaining_count": 15,             // 剩余15份
      "can_purchase": true,              // 可以购买
      "month_price": 9900,               // 99元/月
      // ... 其他字段
    },
    {
      "id": 2,
      "name": "高级套餐",
      "transfer_enable": 1073741824000,  // 1TB
      "speed_limit": 500,
      "device_limit": 5,
      "capacity_limit": null,            // 不限制
      "sold_count": 234,
      "remaining_count": -1,             // -1表示不限制
      "can_purchase": true,
      "month_price": 29900,
      // ... 其他字段
    }
  ]
}
```

### 6. 购买流程

```
1. 用户选择套餐
   ↓
2. 检查 plan.CanPurchase()
   ↓
3. 如果可以购买：
   - 创建订单
   - 订单支付成功后：
     * 更新 user.plan_id
     * plan.sold_count += 1
   ↓
4. 如果不能购买：
   - 返回错误："该套餐已售罄"
```

### 7. 特殊场景处理

#### 场景1：用户更换套餐

```go
// 从套餐A换到套餐B
func ChangePlan(userID, newPlanID int64) error {
    user := GetUser(userID)
    oldPlanID := user.PlanID
    
    // 检查新套餐是否可购买
    newPlan := GetPlan(newPlanID)
    if !newPlan.CanPurchase() {
        return errors.New("套餐已售罄")
    }
    
    // 更新用户套餐
    user.PlanID = newPlanID
    UpdateUser(user)
    
    // 更新计数
    if oldPlanID != nil {
        DecrementPlanSoldCount(*oldPlanID)  // 旧套餐 -1
    }
    IncrementPlanSoldCount(newPlanID)       // 新套餐 +1
    
    return nil
}
```

#### 场景2：用户退订

```go
// 用户退订（可选功能）
func CancelSubscription(userID int64) error {
    user := GetUser(userID)
    if user.PlanID != nil {
        DecrementPlanSoldCount(*user.PlanID)  // 已售数量 -1
        user.PlanID = nil
        UpdateUser(user)
    }
    return nil
}
```

#### 场景3：套餐到期

```go
// 套餐到期不减少 sold_count
// 因为用户可能续费，保持占用名额
// 除非用户主动退订或更换套餐
```

### 8. 管理后台功能

#### 套餐列表显示

```
| 套餐名称 | 流量 | 价格 | 已售/限制 | 状态 |
|---------|------|------|----------|------|
| 基础套餐 | 100GB | ¥99 | 85/100 | 在售 |
| 高级套餐 | 1TB | ¥299 | 234/不限 | 在售 |
| 限量套餐 | 500GB | ¥199 | 50/50 | 售罄 |
```

#### 套餐编辑表单

```jsx
<FormItem label="最大可售数量" name="capacity_limit">
  <InputNumber 
    min={0} 
    placeholder="0或留空表示不限制"
    addonAfter={`已售: ${plan.sold_count}`}
  />
</FormItem>

<FormItem label="流量配额" name="transfer_enable">
  <InputNumber 
    min={0} 
    addonAfter="GB"
    formatter={value => value / 1073741824}
    parser={value => value * 1073741824}
  />
</FormItem>

<FormItem label="速度限制" name="speed_limit">
  <InputNumber min={0} addonAfter="Mbps" />
</FormItem>

<FormItem label="设备限制" name="device_limit">
  <InputNumber min={1} addonAfter="台" />
</FormItem>
```

## 实施步骤

1. ✅ 创建数据库迁移，添加 `sold_count` 字段
2. ✅ 更新 Plan 模型，添加 `SoldCount` 字段
3. ✅ 添加 `CanPurchase()` 和 `GetRemainingCount()` 方法
4. ✅ 修改订单创建逻辑，购买时增加计数
5. ✅ 修改套餐更换逻辑，更新计数
6. ✅ 修改 API 响应，返回购买数量信息
7. ⏳ 前端界面调整

## 优势

1. **限量销售**: 可以创建限量套餐，增加稀缺性
2. **库存管理**: 实时显示剩余数量
3. **自动控制**: 售罄后自动停止销售
4. **营销工具**: "仅剩10个名额"等营销话术
5. **数据统计**: 了解各套餐的销售情况

## 注意事项

1. **并发控制**: 购买时需要加锁，防止超卖
2. **计数准确性**: 定期校验 sold_count 与实际用户数是否一致
3. **历史数据**: 初始化时需要统计现有用户数量
4. **退订策略**: 明确退订是否减少计数
5. **套餐迁移**: 用户更换套餐时正确更新计数
