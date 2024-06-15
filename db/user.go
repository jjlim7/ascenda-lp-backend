package db

import (
	"context"
	"log"
	"time"

	"github.com/vynious/ascenda-lp-backend/types"
)

func CreateUserWithCreateUserRequestBody(ctx context.Context, dbs *DB, userRequestBody types.CreateUserRequestBody, newUUID string) (*types.User, error) {
	var roleID *uint = nil

	if userRequestBody.RoleName != "" {
		role, err := RetrieveRoleWithRoleName(ctx, dbs, userRequestBody.RoleName)
		if err != nil {
			return nil, err
		}
		roleID = &role.Id
	}

	user := types.User{
		Id:        newUUID,
		Email:     userRequestBody.Email,
		FirstName: userRequestBody.FirstName,
		LastName:  userRequestBody.LastName,
		RoleID:    roleID,
		RoleName:  &userRequestBody.RoleName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tx := dbs.Conn.Begin()
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	return &user, tx.Commit().Error
}

func RetrieveUserWithGetUserRequestBody(ctx context.Context, dbs *DB, userRequestBody types.GetUserRequestBody) (*types.User, error) {
	var user types.User
	result := dbs.Conn.WithContext(ctx).Where("email = ?", userRequestBody.Email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func RetrieveUserWithEmail(ctx context.Context, dbs *DB, email string) (*types.User, error) {
	var user types.User
	result := dbs.Conn.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func RetrieveAllUsers(ctx context.Context, dbs *DB) ([]types.User, error) {
	var users []types.User
	result := dbs.Conn.WithContext(ctx).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func DeleteUserWithDeleteUserRequestBody(ctx context.Context, dbs *DB, userRequestBody types.DeleteUserRequestBody) error {
	var user types.User
	res := dbs.Conn.WithContext(ctx).Where("id = ?", userRequestBody.Id).First(&user)
	if res.Error != nil {
		return res.Error
	}

	if err := dbs.Conn.WithContext(ctx).Delete(&user).Error; err != nil {
		return err
	}

	return nil
}

func UpdateUserWithUpdateUserRequestBody(ctx context.Context, dbs *DB, userRequestBody types.UpdateUserRequestBody) (types.User, error) {
	tx := dbs.Conn.Begin()
	log.Println(userRequestBody)
	var user types.User
	if err := tx.Where("id = ?", userRequestBody.Id).First(&user).Error; err != nil {
		tx.Rollback()
		return types.User{}, err
	}

	if userRequestBody.NewFirstName != "" {
		user.FirstName = userRequestBody.NewFirstName
	}

	if userRequestBody.NewLastName != "" {
		user.LastName = userRequestBody.NewLastName
	}

	if userRequestBody.NewRoleName != "" {
		var newRole types.Role
		if err := tx.Where("role_name = ?", userRequestBody.NewRoleName).First(&newRole).Error; err != nil {
			tx.Rollback()
			return types.User{}, err
		}
		user.RoleID = &newRole.Id
		user.RoleName = &newRole.RoleName
	}

	user.UpdatedAt = time.Now()

	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return types.User{}, err
	}

	if err := tx.Commit().Error; err != nil {
		return types.User{}, err
	}

	return user, nil

}
