package auth

import (
	authapi "imanager/pkg/api/auth"
	authdb "imanager/pkg/db/auth"
)

func transformUserDB2API(in authdb.User) authapi.User {
	res := authapi.User{
		UUID:      in.UUID,
		Name:      in.Name,
		Password:  in.Password,
		TruthName: in.TruthName,
		Email:     in.Email,
		PhoneNum:  in.PhoneNum,
		Role: make([]authapi.RoleInUser, 0, len(in.Role)),
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
		Role: make([]*authdb.Role, 0, len(in.Role)),
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
	return authapi.Group{
		ID:         in.Id,
		Name:       in.Name,
		Annotation: in.Annotation,
	}
}

func transformGroupAPI2DB(in authapi.Group) authdb.Group {
	return authdb.Group{
		Id:         in.ID,
		Name:       in.Name,
		Annotation: in.Annotation,
	}
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
	}
}

func transformRoleAPI2DB(in authapi.Role) authdb.Role {
	return authdb.Role{
		Id:         in.ID,
		Name:       in.Name,
		Annotation: in.Annotation,
	}
}

func transformRoleDBs2APIs(in []authdb.Role) []authapi.Role {
	res := make([]authapi.Role, 0, len(in))
	for _, v := range in {
		res = append(res, transformRoleDB2API(v))
	}
	return res
}
