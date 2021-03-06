package messaging

import (
	"encoding/json"
	"fmt"
	"github.com/cortezaproject/corteza-server/messaging/rest/request"
	"github.com/cortezaproject/corteza-server/messaging/types"
	"github.com/cortezaproject/corteza-server/pkg/id"
	"github.com/cortezaproject/corteza-server/tests/helpers"
	"github.com/steinfletcher/apitest"
	"github.com/steinfletcher/apitest-jsonpath"
	"net/http"
	"strconv"
	"testing"
)

func (h helper) chUpdate(ch *request.ChannelUpdate) *apitest.Response {
	payload, err := json.Marshal(ch)
	h.a.NoError(err)

	return h.apiInit().
		Put(fmt.Sprintf("/channels/%v", ch.ChannelID)).
		Header("Accept", "application/json").
		JSON(string(payload)).
		Expect(h.t).
		Status(http.StatusOK)
}

func channelToRequest(ch *types.Channel) *request.ChannelUpdate {
	req := &request.ChannelUpdate{
		ChannelID:        ch.ID,
		Name:             ch.Name,
		Topic:            ch.Topic,
		MembershipPolicy: ch.MembershipPolicy,
		Type:             ch.Type.String(),
	}

	return req
}

func TestChannelUpdateNonexistent(t *testing.T) {
	h := newHelper(t)

	req := &request.ChannelUpdate{
		ChannelID: id.Next(),
		Name:      "some name",
		Type:      "public",
	}

	h.chUpdate(req).
		Assert(helpers.AssertError("channel does not exist")).
		End()

}

func TestChannelUpdateDenied(t *testing.T) {
	h := newHelper(t)
	ch := h.repoMakePublicCh()

	h.deny(ch.RBACResource(), "update")

	req := channelToRequest(ch)
	req.Name = "Updated name"

	h.chUpdate(req).
		Assert(helpers.AssertError("not allowed to update this channel")).
		End()
}

func TestChannelUpdate(t *testing.T) {
	h := newHelper(t)
	ch := h.repoMakePublicCh()

	h.allow(ch.RBACResource(), "update")

	req := channelToRequest(ch)
	req.Name = "Updated name"

	h.chUpdate(req).
		Assert(helpers.AssertNoErrors).
		Assert(jsonpath.Equal(`$.response.name`, req.Name)).
		Assert(jsonpath.Equal(`$.response.channelID`, strconv.FormatUint(ch.ID, 10))).
		End()
}
