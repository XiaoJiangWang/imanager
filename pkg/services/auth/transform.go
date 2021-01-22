package auth

import (
	authapi "imanager/pkg/api/auth"
	apiutil "imanager/pkg/api/util"
	authdb "imanager/pkg/db/auth"
	dbutil "imanager/pkg/db/util"
)

func transformUserDB2API(in authdb.User) authapi.User {
	res := authapi.User{
		UUID:      in.UUID,
		Name:      in.Name,
		Password:  in.Password,
		TruthName: in.TruthName,
		Email:     in.Email,
		PhoneNum:  in.PhoneNum,
		Role:      make([]authapi.RoleInUser, 0, len(in.Role)),
		BaseModel: apiutil.BaseModel{
			CreateTimestamp: in.CreateTimestamp,
			UpdateTimestamp: in.UpdateTimestamp,
		},
	}
	if in.Group != nil {
		res.Group = &authapi.GroupInUser{
			ID:         in.Group.Id,
			Name:       in.Group.Name,
			Annotation: in.Group.Annotation,
		}
	}
	for _, v := range in.Role {
		res.Role = append(res.Role, authapi.RoleInUser{
			ID:         v.Id,
			Name:       v.Name,
			Annotation: v.Annotation,
		})
	}
	return res
}

func transformUserAPI2DB(in authapi.User) authdb.User {
	res := authdb.User{
		UUID:      in.UUID,
		Name:      in.Name,
		Password:  in.Password,
		TruthName: in.TruthName,
		Email:     in.Email,
		PhoneNum:  in.PhoneNum,
		Role:      make([]*authdb.Role, 0, len(in.Role)),
		BaseModel: dbutil.BaseModel{
			CreateTimestamp: in.CreateTimestamp,
			UpdateTimestamp: in.UpdateTimestamp,
		},
	}
	if in.Group != nil {
		res.Group = &authdb.Group{
			Id:         in.Group.ID,
			Name:       in.Group.Name,
			Annotation: in.Group.Annotation,
		}
	}
	for _, v := range in.Role {
		res.Role = append(res.Role, &authdb.Role{
			Id:         v.ID,
			Name:       v.Name,
			Annotation: v.Annotation,
		})
	}
	return res
}

func transformUserDBs2APIs(input []authdb.User) []authapi.User {
	res := make([]authapi.User, 0, len(input))
	for _, v := range input {
		res = append(res, transformUserDB2API(v))
	}
	return res
}

func transformGroupDB2API(in authdb.Group) authapi.Group {
	res := authapi.Group{
		ID:         in.Id,
		Name:       in.Name,
		Annotation: in.Annotation,
		Builtin:    in.Builtin,
		BaseModel: apiutil.BaseModel{
			CreateTimestamp: in.CreateTimestamp,
			UpdateTimestamp: in.UpdateTimestamp,
		},
	}
	if in.Role != nil && len(in.Role) != 0 {
		res.Role = make([]authapi.RoleInGroup, 0, len(in.Role))
		for _, role := range in.Role {
			res.Role = append(res.Role, authapi.RoleInGroup{
				ID:         role.Id,
				Name:       role.Name,
				Annotation: role.Annotation,
			})
		}
	}
	if in.User != nil && len(in.User) != 0 {
		res.User = make([]authapi.UserInGroup, 0, len(in.User))
		for _, user := range in.User {
			res.User = append(res.User, authapi.UserInGroup{
				ID:   user.ID,
				UUID: user.UUID,
				Name: user.Name,
			})
		}
	}
	return res
}

func transformGroupAPI2DB(in authapi.Group) authdb.Group {
	res := authdb.Group{
		Id:         in.ID,
		Name:       in.Name,
		Annotation: in.Annotation,
		Builtin:    in.Builtin,
		BaseModel: dbutil.BaseModel{
			CreateTimestamp: in.CreateTimestamp,
			UpdateTimestamp: in.UpdateTimestamp,
		},
	}
	for in.User != nil && len(in.User) != 0 {
		res.User = make([]*authdb.User, 0, len(in.User))
		for _, user := range in.User {
			res.User = append(res.User, &authdb.User{
				ID:   user.ID,
				UUID: user.UUID,
				Name: user.Name,
			})
		}
	}
	for in.Role != nil && len(in.Role) != 0 {
		res.Role = make([]*authdb.Role, 0, len(in.Role))
		for _, role := range in.Role {
			res.Role = append(res.Role, &authdb.Role{
				Id:         role.ID,
				Name:       role.Name,
				Annotation: role.Annotation,
			})
		}
	}
	return res
}

func transformGroupDBs2APIs(in []authdb.Group) []authapi.Group {
	res := make([]authapi.Group, 0, len(in))
	for _, v := range in {
		res = append(res, transformGroupDB2API(v))
	}
	return res
}

func transformRoleDB2API(in authdb.Role) authapi.Role {
	return authapi.Role{
		ID:         in.Id,
		Name:       in.Name,
		Annotation: in.Annotation,
		BaseModel: apiutil.BaseModel{
			CreateTimestamp: in.CreateTimestamp,
			UpdateTimestamp: in.UpdateTimestamp,
		},
	}
}

func transformRoleAPI2DB(in authapi.Role) authdb.Role {
	return authdb.Role{
		Id:         in.ID,
		Name:       in.Name,
		Annotation: in.Annotation,
		BaseModel: dbutil.BaseModel{
			CreateTimestamp: in.CreateTimestamp,
			UpdateTimestamp: in.UpdateTimestamp,
		},
	}
}

func transformRoleDBs2APIs(in []authdb.Role) []authapi.Role {
	res := make([]authapi.Role, 0, len(in))
	for _, v := range in {
		res = append(res, transformRoleDB2API(v))
	}
	return res
}
