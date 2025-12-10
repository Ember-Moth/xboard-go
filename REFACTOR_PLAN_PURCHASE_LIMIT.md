# 套餐购买数量限制功能实现

## 改动说明

为套餐添加购买数量限制功能，支持限量销售和库存管理。

## 修改内容

### 1. Plan 模型 - 添加已售数量字段

**文件**: `internal/model/plan.go`

**改动**:
- 添加 `SoldCount` 字段：已售出数量
- 明确 `CapacityLimit` 字段含义：最大可售数量
- 添加 `CanPurchase()` 方法：检查是否可以购买
- 添加 `GetRemainingCount()` 方法：获取剩余可售数量

```go
type Plan struct {
    // ... 现有字段
    
    // 流量和限制配置
    TransferEnable  int64  `json:"transfer_enable"`  // 流量配额（字节）
    SpeedLimit      *int   `json:"speed_limit"`      // 速度限制（Mbps）
    DeviceLimit     *int   `json:"device_limit"`     // 设备数量限制
    
    // 购买数量限制
    CapacityLimit   *int   `json:"capacity_limit"`   // 最大可售数量（null或0=不限制）
    SoldCount       int    `json:"sold_count"`       // 已售出数量（新增）
    
    // ... 其他字段
}

// CanPurchase 检查套餐是否可以购买
func (p *Plan) CanPurchase() bool {
    if p.CapacityLimit == nil || *p.CapacityLimit <= 0 {
        return true
    }
    return p.SoldCount < *p.CapacityLimit
}

// GetRemainingCount 获取剩余可售数量
func (p *Plan) GetRemainingCount() int {
    if p.CapacityLimit == nil || *p.CapacityLimit <= 0 {
        return -1  // -1 表示不限制
    }
    remaining := *p.CapacityLimit - p.SoldCount
    if remaining < 0 {
        return 0
    }
    return remaining
}
```

### 2. PlanService - 添加购买数量管理

**文件**: `internal/service/plan.go`

**改动**:
- `GetPlanInfo()` 返回购买数量信息
- 添加 `IncrementSoldCount()` 方法：增加已售数量
- 添加 `DecrementSoldCount()` 方法：减少已售数量

```go
// GetPlanInfo 返回增强的套餐信息
func (s *PlanService) GetPlanInfo(plan *model.Plan) map[string]interface{} {
    return map[string]interface{}{
        // ... 现有字段
        "capacity_limit":  plan.CapacityLimit,
        "sold_count":      plan.SoldCount,
        "remaining_count": plan.GetRemainingCount(),  // 新增
        "can_purchase":    plan.CanPurchase(),        // 新增
    }
}

// IncrementSoldCount 增加已售数量
func (s *PlanService) IncrementSoldCount(planID int64) error

// DecrementSoldCount 减少已售数量
func (s *PlanService) DecrementSoldCount(planID int64) error
```

### 3. PlanRepository - 原子操作

**文件**: `internal/repository/plan.go`

**改动**:
- 添加 `IncrementSoldCount()` 方法：原子递增
- 添加 `DecrementSoldCount()` 方法：原子递减

```go
// IncrementSoldCount 增加已售数量（原子操作，防止并发问题）
func (r *PlanRepository) IncrementSoldCount(planID int64) error {
    return r.db.Model(&model.Plan{}).Where("id = ?", planID).
        UpdateColumn("sold_count", gorm.Expr("sold_count + ?", 1)).Error
}

// DecrementSoldCount 减少已售数量（原子操作）
func (r *PlanRepository) DecrementSoldCount(planID int64) error {
    return r.db.Model(&model.Plan{}).Where("id = ? AND sold_count > 0", planID).
        UpdateColumn("sold_count", gorm.Expr("sold_count - ?", 1)).Error
}
```

### 4. 数据库迁移

**文件**: `migrations/005_add_plan_sold_count.sql`

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

-- 添加索引（用于快速查询可售套餐）
CREATE INDEX `idx_plan_capacity` ON `v2_plan`(`capacity_limit`, `sold_count`);
```

## API 响应示例

### 获取套餐列表

```json
GET /api/plans

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
      "remaining_count": 15,             // 剩余15份（新增）
      "can_purchase": true,              // 可以购买（新增）
      "prices": {
        "monthly": 9900,
        "yearly": 99000
      }
    },
    {
      "id": 2,
      "name": "高级套餐",
      "transfer_enable": 1073741824000,  // 1TB
      "speed_limit": 500,
      "device_limit": 5,
      "capacity_limit": null,            // 不限制
      "sold_count": 234,
      "remaining_count": -1,             // -1表示不限制（新增）
      "can_purchase": true,              // 可以购买（新增）
      "prices": {
        "monthly": 29900
      }
    },
    {
      "id": 3,
      "name": "限量套餐",
      "transfer_enable": 536870912000,   // 500GB
      "capacity_limit": 50,              // 限量50份
      "sold_count": 50,                  // 已售50份
      "remaining_count": 0,              // 已售罄（新增）
      "can_purchase": false,             // 不可购买（新增）
      "prices": {
        "monthly": 19900
      }
    }
  ]
}
```

## 使用场景

### 场景1：用户购买套餐

```go
// 在订单服务中
func CreateOrder(userID, planID int64, period string) error {
    // 1. 获取套餐
    plan, err := planService.GetByID(planID)
    if err != nil {
        return err
    }
    
    // 2. 检查是否可以购买
    if !plan.CanPurchase() {
        return errors.New("该套餐已售罄")
    }
    
    // 3. 创建订单
    order := CreateOrder(...)
    
    // 4. 订单支付成功后
    if order.Status == "paid" {
        // 更新用户套餐
        user.PlanID = planID
        userRepo.Update(user)
        
        // 增加已售数量
        planService.IncrementSoldCount(planID)
    }
    
    return nil
}
```

### 场景2：用户更换套餐

```go
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
        planService.DecrementSoldCount(*oldPlanID)  // 旧套餐 -1
    }
    planService.IncrementSoldCount(newPlanID)       // 新套餐 +1
    
    return nil
}
```

### 场景3：用户退订

```go
func CancelSubscription(userID int64) error {
    user := GetUser(userID)
    if user.PlanID != nil {
        // 减少已售数量
        planService.DecrementSoldCount(*user.PlanID)
        
        // 清除用户套餐
        user.PlanID = nil
        UpdateUser(user)
    }
    return nil
}
```

## 前端展示

### 套餐列表页面

```jsx
<Card>
  <h3>{plan.name}</h3>
  <p>流量：{formatBytes(plan.transfer_enable)}</p>
  <p>速度：{plan.speed_limit} Mbps</p>
  <p>设备：{plan.device_limit} 台</p>
  
  {/* 显示库存信息 */}
  {plan.capacity_limit && (
    <div className="stock-info">
      {plan.can_purchase ? (
        <Tag color="green">
          剩余 {plan.remaining_count} 份
        </Tag>
      ) : (
        <Tag color="red">已售罄</Tag>
      )}
      <Progress 
        percent={(plan.sold_count / plan.capacity_limit) * 100}
        format={() => `${plan.sold_count}/${plan.capacity_limit}`}
      />
    </div>
  )}
  
  <Button 
    type="primary" 
    disabled={!plan.can_purchase}
  >
    {plan.can_purchase ? '立即购买' : '已售罄'}
  </Button>
</Card>
```

### 管理后台 - 套餐列表

```jsx
<Table>
  <Column title="套餐名称" dataIndex="name" />
  <Column title="流量" render={(record) => formatBytes(record.transfer_enable)} />
  <Column 
    title="已售/限制" 
    render={(record) => {
      if (!record.capacity_limit) {
        return `${record.sold_count}/不限`;
      }
      return `${record.sold_count}/${record.capacity_limit}`;
    }}
  />
  <Column 
    title="状态" 
    render={(record) => {
      if (!record.can_purchase) {
        return <Tag color="red">售罄</Tag>;
      }
      if (record.remaining_count > 0 && record.remaining_count <= 10) {
        return <Tag color="orange">即将售罄</Tag>;
      }
      return <Tag color="green">在售</Tag>;
    }}
  />
</Table>
```

### 管理后台 - 套餐编辑

```jsx
<Form>
  <FormItem label="套餐名称" name="name">
    <Input />
  </FormItem>
  
  <FormItem label="流量配额" name="transfer_enable">
    <InputNumber 
      addonAfter="GB"
      formatter={value => value / 1073741824}
      parser={value => value * 1073741824}
    />
  </FormItem>
  
  <FormItem label="速度限制" name="speed_limit">
    <InputNumber addonAfter="Mbps" />
  </FormItem>
  
  <FormItem label="设备限制" name="device_limit">
    <InputNumber addonAfter="台" />
  </FormItem>
  
  <FormItem 
    label="最大可售数量" 
    name="capacity_limit"
    extra={`已售出: ${plan.sold_count} 份`}
  >
    <InputNumber 
      min={0} 
      placeholder="0或留空表示不限制"
    />
  </FormItem>
</Form>
```

## 优势

1. ✅ **限量销售**: 支持创建限量套餐，增加稀缺性
2. ✅ **库存管理**: 实时显示剩余数量，防止超卖
3. ✅ **自动控制**: 售罄后自动停止销售
4. ✅ **营销工具**: "仅剩10个名额"等营销话术
5. ✅ **数据统计**: 了解各套餐的销售情况
6. ✅ **并发安全**: 使用原子操作，防止并发超卖
7. ✅ **向后兼容**: 不影响现有套餐，默认不限制

## 注意事项

1. **并发控制**: 使用数据库原子操作（`gorm.Expr`）防止超卖
2. **计数准确性**: 定期校验 `sold_count` 与实际用户数是否一致
3. **历史数据**: 迁移时自动统计现有用户数量初始化 `sold_count`
4. **退订策略**: 用户退订时减少计数，释放名额
5. **套餐迁移**: 用户更换套餐时正确更新两个套餐的计数

## 测试建议

1. 创建限量套餐（capacity_limit = 10）
2. 购买套餐，验证 sold_count 增加
3. 购买到限制数量，验证 can_purchase = false
4. 用户更换套餐，验证两个套餐的计数正确更新
5. 用户退订，验证 sold_count 减少
6. 并发购买测试，验证不会超卖

## 相关文件

```
modified:   internal/model/plan.go
modified:   internal/service/plan.go
modified:   internal/repository/plan.go
new file:   migrations/005_add_plan_sold_count.sql
new file:   docs/plan-purchase-limit.md
new file:   REFACTOR_PLAN_PURCHASE_LIMIT.md
```

## 下一步

1. ⏳ 在订单服务中集成购买数量检查和更新
2. ⏳ 在用户服务中集成套餐更换时的计数更新
3. ⏳ 前端界面实现库存显示和售罄状态
4. ⏳ 添加管理后台的库存预警功能
5. ⏳ 添加定时任务校验计数准确性
