package types

import (
	"time"
)

type Role struct {
	Id           uint               `gorm:"primaryKey"`
	RoleName     string             `gorm:"unique"`
	Users        []*User            `gorm:"foreignKey:RoleID"` // new
	Permissions  RolePermissionList `gorm:"foreignKey:RoleID"`
	MakerRoles   []ApprovalChainMap `gorm:"foreignKey:MakerRoleID"`   // new
	CheckerRoles []ApprovalChainMap `gorm:"foreignKey:CheckerRoleID"` // new
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RolePermission struct {
	Id        uint   `gorm:"primaryKey"`
	RoleID    uint   `gorm:"index"`
	CanCreate bool   `gorm:"default:false"`
	CanRead   bool   `gorm:"default:false"`
	CanUpdate bool   `gorm:"default:false"`
	CanDelete bool   `gorm:"default:false"`
	Resource  string `json:"resource"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RoleList []Role

type RolePermissionList []RolePermission

type CreateRoleRequestBody struct {
	RoleName    string
	Permissions *RolePermissionList
}

type GetRoleRequestBody struct {
	RoleName string
}

type DeleteRoleRequestBody struct {
	RoleName string
}

type UpdateRoleRequestBody struct {
	RoleName    string
	NewRoleName string
	Permissions *RolePermissionList
}
