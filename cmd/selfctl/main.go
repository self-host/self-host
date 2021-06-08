/*
Copyright (C) 2021 The Self-host Authors.
This file is part of Self-host <https://github.com/self-host/self-host>.

Self-host is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Self-host is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with dogtag.  If not, see <http://www.gnu.org/licenses/>.
*/
package main

import (
	"context"
	"fmt"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"

	"github.com/noda/selfhost/api/v2/rest"
)

func main() {
	basicAuthProvider, basicAuthProviderErr := securityprovider.NewSecurityProviderBasicAuth("test", "root")
	if basicAuthProviderErr != nil {
		panic(basicAuthProviderErr)
	}

	client, err := rest.NewClient("http://127.0.0.1:8095/", rest.WithRequestEditorFn(basicAuthProvider.Intercept))
	if err != nil {
	}

	limit := rest.LimitParam(int64(100))
	offset := rest.OffsetParam(int64(0))

	params := &rest.FindUsersParams{
		Limit:  &limit,
		Offset: &offset,
	}

	req, err := client.FindUsers(context.Background(), params)

	fmt.Println(req)
}
