package dao

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("邮箱冲突")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (dao *UserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	if me, ok := err.(*mysql.MySQLError); ok {
		const duplicateErr uint16 = 1062
		if me.Number == duplicateErr {
			// 用户冲突，邮箱冲突
			return ErrDuplicateEmail
		}
	}
	return err
}

func (dao *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email=?", email).First(&u).Error
	return u, err
}

func (dao *UserDAO) ModifyAuthTimesByEmail(ctx context.Context, email string, auth_hide_times int, auth_extract_times int) error {
	result := dao.db.WithContext(ctx).Model(&User{}).Where("email=?", email).Updates(map[string]interface{}{
		"Remain_hide_times":    auth_hide_times,
		"Remain_extract_times": auth_extract_times,
	})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (dao *UserDAO) HideMinusOneByUserId(ctx context.Context, userId int64) (int, error) {

	// 更新值
	//result := dao.db.Raw("UPDATE users SET remain_hide_times = remain_hide_times - 1 WHERE id = ?", userId)

	result := dao.db.WithContext(ctx).Table("users").Where("id = ?", userId).Update("remain_hide_times", gorm.Expr("remain_hide_times - ?", 1))

	if result.Error != nil {
		fmt.Println(result.Error.Error())
		return 0, result.Error
	}

	// 查询值
	var updatedTimes int
	err := dao.db.WithContext(ctx).Raw("SELECT remain_hide_times FROM users WHERE id = ?", userId).Scan(&updatedTimes).Error
	if err != nil {
		return 0, err
	}
	return updatedTimes, nil
}

func (dao *UserDAO) ExtractMinusOneByUserId(ctx context.Context, userId int64) (int, error) {
	// 更新值
	//err := dao.db.WithContext(ctx).Raw("UPDATE users SET remain_extract_times = remain_hide_times - 1 WHERE id = ?", userId).Error

	result := dao.db.WithContext(ctx).Table("users").Where("id = ?", userId).Update("remain_extract_times", gorm.Expr("remain_extract_times - ?", 1))

	if result.Error != nil {
		fmt.Println(result.Error.Error())
		return 0, result.Error
	}

	// 查询值
	var updatedTimes int
	err := dao.db.WithContext(ctx).Raw("SELECT remain_extract_times FROM users WHERE id = ?", userId).Scan(&updatedTimes).Error
	if err != nil {
		return 0, err
	}
	return updatedTimes, nil
}

func (dao *UserDAO) HideCheckByUserId(ctx context.Context, userId int64) (int, error) {
	var remainTimes int

	// 使用原生SQL执行更新并返回更新后的值
	err := dao.db.WithContext(ctx).Raw("SELECT remain_hide_times FROM users WHERE id = ?", userId).Scan(&remainTimes).Error

	if err != nil {
		return 0, err
	}

	return remainTimes, nil
}

func (dao *UserDAO) ExtractCheckByUserId(ctx context.Context, userId int64) (int, int, error) {
	//var hideRemainTimes int
	//var extractRemainTimes int
	type UserRemains struct {
		RemainHideTimes    int `gorm:"column:remain_hide_times"`
		RemainExtractTimes int `gorm:"column:remain_extract_times"`
	}
	var r UserRemains
	// 使用原生SQL执行更新并返回更新后的值
	err := dao.db.WithContext(ctx).Raw("SELECT remain_hide_times,remain_extract_times FROM users WHERE id = ?", userId).Scan(&r).Error

	if err != nil {
		return 0, 0, err
	}

	return r.RemainHideTimes, r.RemainExtractTimes, nil
}

type User struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	Email    string `gorm:"unique"`
	Password string

	// 创建时间
	Ctime int64
	// 更新时间
	Utime int64

	// 隐藏次数
	Remain_hide_times int
	// 提取次数
	Remain_extract_times int
}
