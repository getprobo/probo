// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package console_v1

import (
	"context"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
)

// scheduleDeviceSync makes a device-sync connector's devices sync on the
// worker's next poll. The connector is already saved, so a failure here is
// logged rather than failing the mutation: the worker creates the sync of any
// connector left without one on its own next poll.
func (r *mutationResolver) scheduleDeviceSync(
	ctx context.Context,
	scope coredata.Scoper,
	cnnctr *coredata.Connector,
) {
	reg, ok := r.providerRegistry.Get(cnnctr.Provider)
	if !ok || !reg.SyncsDevices() {
		return
	}

	if err := r.itam.ScheduleDeviceSync(ctx, scope, cnnctr); err != nil {
		r.logger.ErrorCtx(
			ctx,
			"cannot schedule device sync",
			log.String("connector_id", cnnctr.ID.String()),
			log.Error(err),
		)
	}
}
