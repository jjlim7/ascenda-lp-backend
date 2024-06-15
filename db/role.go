package db

import (
	"context"
	"time"

	"github.com/vynious/ascenda-lp-backend/types"
	"gorm.io/gorm"
)

func CreateRoleWithCreateRoleRequestBody(ctx context.Context, dbs *DB, roleRequestBody types.CreateRoleRequestBody) (string, error) {
	role := types.Role{
		RoleName:  roleRequestBody.RoleName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if roleRequestBody.Permissions != nil {
		role.Permissions = *roleRequestBody.Permissions
	}

	tx := dbs.Conn.Begin()
	if err := tx.Create(&role).Error; err != nil {
		tx.Rollback()
		return "", err
	}

	return role.RoleName, tx.Commit().Error
}

func RetrieveRoleWithRoleName(ctx context.Context, dbs *DB, roleName string) (types.Role, error) {
	var role types.Role
	if err := dbs.Conn.Preload("Permissions").Preload("Users").Where("role_name = ?", roleName).First(&role).Error; err != nil {
		return types.Role{}, err
	}
	return role, nil
}

func RetrieveAllRolesWithUsers(ctx context.Context, dbs *DB) ([]types.Role, error) {
	var roles []types.Role
	if result := dbs.Conn.WithContext(ctx).Preload("Permissions").Preload("Users").Find(&roles); result.Error != nil {
		return nil, result.Error
	}
	return roles, nil
}

func RetrieveRoleWithRetrieveRoleRequestBody(ctx context.Context, dbs *DB, roleRequestBody types.GetRoleRequestBody) (types.Role, error) {
	var role types.Role
	if err := dbs.Conn.Preload("Permissions").Preload("Users").Where("role_name = ?", roleRequestBody.RoleName).First(&role).Error; err != nil {
		return types.Role{}, err
	}
	return role, nil
}

func DeleteRoleWithDeleteRoleRequestBody(ctx context.Context, dbs *DB, roleRequestBody types.DeleteRoleRequestBody) error {
	tx := dbs.Conn.Begin()

	var role types.Role
	if err := tx.Where("role_name = ?", roleRequestBody.RoleName).First(&role).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&types.User{}).Where("role_id = ?", role.Id).Update("role_id", gorm.Expr("NULL")).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("role_id = ?", role.Id).Delete(&types.RolePermission{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&role).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func UpdateRole(ctx context.Context, dbs *DB, roleRequestBody types.UpdateRoleRequestBody) (types.Role, error) {
	tx := dbs.Conn.Begin()

	var role types.Role
	if err := tx.Where("role_name = ?", roleRequestBody.RoleName).First(&role).Error; err != nil {
		tx.Rollback()
		return types.Role{}, err
	}

	if roleRequestBody.NewRoleName != "" {
		role.RoleName = roleRequestBody.NewRoleName
		role.UpdatedAt = time.Now()
		if err := tx.Save(&role).Error; err != nil {
			tx.Rollback()
			return types.Role{}, err
		}
	}

	if roleRequestBody.Permissions != nil {
		if err := tx.Where("role_id = ?", role.Id).Delete(&types.RolePermission{}).Error; err != nil {
			tx.Rollback()
			return types.Role{}, err
		}

		for _, perm := range *roleRequestBody.Permissions {
			perm.RoleID = role.Id
			if err := tx.Create(&perm).Error; err != nil {
				tx.Rollback()
				return types.Role{}, err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return types.Role{}, err
	}
	if err := dbs.Conn.Preload("Permissions").Preload("Users").Where("id = ?", role.Id).First(&role).Error; err != nil {
		return types.Role{}, err
	}

	return role, nil
}
