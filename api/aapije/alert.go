// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package aapije

import (
	"net/http"

	"github.com/self-host/self-host/api/aapije/rest"
)

// List alerts
func (ra *RestApi) FindAlerts(w http.ResponseWriter, r *http.Request, p rest.FindAlertsParams) {
	w.WriteHeader(http.StatusNotImplemented)
}
