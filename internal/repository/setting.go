package repository

import (
	"xboard/internal/model"

	"gorm.io/gorm"
)

type SettingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) *SettingRepository {
	return &SettingRepository{db: db}
}

func (r *SettingRepository) Get(key string) (string, error) {
	var setting model.Setting
	err := r.db.Where("`key` = ?", key).First(&setting).Error
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (r *SettingRepository) Set(key, value string) error {
	var setting model.Setting
	err := r.db.Where("`key` = ?", key).First(&setting).Error
	if err == gorm.ErrRecordNotFound {
		setting = model.Setting{Key: key, Value: value}
		return r.db.Create(&setting).Error
	}
	if err != nil {
		return err
	}
	setting.Value = value
	return r.db.Save(&setting).Error
}

func (r *SettingRepository) GetAll() (map[string]string, error) {
	var settings []model.Setting
	err := r.db.Find(&settings).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result, nil
}

func (r *SettingRepository) Delete(key string) error {
	return r.db.Where("`key` = ?", key).Delete(&model.Setting{}).Error
}


// NoticeRepository 公告仓库
type NoticeRepository struct {
	db *gorm.DB
}

func NewNoticeRepository(db *gorm.DB) *NoticeRepository {
	return &NoticeRepository{db: db}
}

func (r *NoticeRepository) Create(notice *model.Notice) error {
	return r.db.Create(notice).Error
}

func (r *NoticeRepository) Update(notice *model.Notice) error {
	return r.db.Save(notice).Error
}

func (r *NoticeRepository) Delete(id int64) error {
	return r.db.Delete(&model.Notice{}, id).Error
}

func (r *NoticeRepository) FindByID(id int64) (*model.Notice, error) {
	var notice model.Notice
	err := r.db.First(&notice, id).Error
	if err != nil {
		return nil, err
	}
	return &notice, nil
}

func (r *NoticeRepository) GetAll() ([]model.Notice, error) {
	var notices []model.Notice
	err := r.db.Order("sort ASC, id DESC").Find(&notices).Error
	return notices, err
}

func (r *NoticeRepository) GetVisible() ([]model.Notice, error) {
	var notices []model.Notice
	err := r.db.Where("`show` = ?", true).Order("sort ASC, id DESC").Find(&notices).Error
	return notices, err
}

// KnowledgeRepository 知识库仓库
type KnowledgeRepository struct {
	db *gorm.DB
}

func NewKnowledgeRepository(db *gorm.DB) *KnowledgeRepository {
	return &KnowledgeRepository{db: db}
}

func (r *KnowledgeRepository) Create(knowledge *model.Knowledge) error {
	return r.db.Create(knowledge).Error
}

func (r *KnowledgeRepository) Update(knowledge *model.Knowledge) error {
	return r.db.Save(knowledge).Error
}

func (r *KnowledgeRepository) Delete(id int64) error {
	return r.db.Delete(&model.Knowledge{}, id).Error
}

func (r *KnowledgeRepository) FindByID(id int64) (*model.Knowledge, error) {
	var knowledge model.Knowledge
	err := r.db.First(&knowledge, id).Error
	if err != nil {
		return nil, err
	}
	return &knowledge, nil
}

func (r *KnowledgeRepository) GetAll() ([]model.Knowledge, error) {
	var items []model.Knowledge
	err := r.db.Order("sort ASC, id DESC").Find(&items).Error
	return items, err
}

func (r *KnowledgeRepository) GetVisible(language string) ([]model.Knowledge, error) {
	var items []model.Knowledge
	query := r.db.Where("`show` = ?", true)
	if language != "" {
		query = query.Where("language = ?", language)
	}
	err := query.Order("sort ASC, id DESC").Find(&items).Error
	return items, err
}

func (r *KnowledgeRepository) GetByCategory(category, language string) ([]model.Knowledge, error) {
	var items []model.Knowledge
	query := r.db.Where("`show` = ? AND category = ?", true, category)
	if language != "" {
		query = query.Where("language = ?", language)
	}
	err := query.Order("sort ASC, id DESC").Find(&items).Error
	return items, err
}

func (r *KnowledgeRepository) GetCategories(language string) ([]string, error) {
	var categories []string
	query := r.db.Model(&model.Knowledge{}).Where("`show` = ?", true)
	if language != "" {
		query = query.Where("language = ?", language)
	}
	err := query.Distinct("category").Pluck("category", &categories).Error
	return categories, err
}

// InviteCodeRepository 邀请码仓库
type InviteCodeRepository struct {
	db *gorm.DB
}

func NewInviteCodeRepository(db *gorm.DB) *InviteCodeRepository {
	return &InviteCodeRepository{db: db}
}

func (r *InviteCodeRepository) Create(code *model.InviteCode) error {
	return r.db.Create(code).Error
}

func (r *InviteCodeRepository) Update(code *model.InviteCode) error {
	return r.db.Save(code).Error
}

func (r *InviteCodeRepository) FindByCode(code string) (*model.InviteCode, error) {
	var inviteCode model.InviteCode
	err := r.db.Where("code = ?", code).First(&inviteCode).Error
	if err != nil {
		return nil, err
	}
	return &inviteCode, nil
}

func (r *InviteCodeRepository) FindByUserID(userID int64) ([]model.InviteCode, error) {
	var codes []model.InviteCode
	err := r.db.Where("user_id = ?", userID).Order("id DESC").Find(&codes).Error
	return codes, err
}

// CommissionLogRepository 佣金记录仓库
type CommissionLogRepository struct {
	db *gorm.DB
}

func NewCommissionLogRepository(db *gorm.DB) *CommissionLogRepository {
	return &CommissionLogRepository{db: db}
}

func (r *CommissionLogRepository) Create(log *model.CommissionLog) error {
	return r.db.Create(log).Error
}

func (r *CommissionLogRepository) FindByUserID(userID int64, page, pageSize int) ([]model.CommissionLog, int64, error) {
	var logs []model.CommissionLog
	var total int64

	r.db.Model(&model.CommissionLog{}).Where("invite_user_id = ?", userID).Count(&total)
	err := r.db.Where("invite_user_id = ?", userID).
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}

func (r *CommissionLogRepository) SumByUserID(userID int64) (int64, error) {
	var total int64
	err := r.db.Model(&model.CommissionLog{}).
		Where("invite_user_id = ?", userID).
		Select("COALESCE(SUM(get_amount), 0)").
		Scan(&total).Error
	return total, err
}
