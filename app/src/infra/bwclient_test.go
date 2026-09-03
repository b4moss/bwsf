package infra

import (
	"errors"
	"os"
	"testing"

	"bwsf/src/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const infraTestSession = "P4tHpDULkFR5+NLL1lbfxD43q9NqIS2tmKnG0GMAn/Ft8w4JOipXty4uY4EQ5/gkTXDPGpidXuoC155F65X5sQ=="

func stubUtilsBw(t *testing.T, handler func(utils.TestBwCall) utils.TestBwReply) {
	t.Helper()
	t.Setenv("NO_COLOR", "1")
	restore := utils.InstallBwExecTestHook(
		func(file string) (string, error) {
			if file == "bw" {
				return "/mock/bw", nil
			}
			return "", errors.New("not found")
		},
		handler,
	)
	t.Cleanup(func() {
		restore()
		_ = os.Unsetenv("BW_SESSION")
	})
}

func TestRealBwClient_ListItemsInFolder_Converts(t *testing.T) {
	stubUtilsBw(t, func(call utils.TestBwCall) utils.TestBwReply {
		assert.Equal(t, []string{"list", "items", "--folderid", "fid"}, call.Args)
		return utils.TestBwReply{Output: []byte(`[{"id":"i1","name":"proj"}]`)}
	})
	c := NewBwClient()
	items, err := c.ListItemsInFolder("fid")
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "i1", items[0].ID)
	assert.Equal(t, "proj", items[0].Name)
}

func TestRealBwClient_GetItemByID_Converts(t *testing.T) {
	stubUtilsBw(t, func(call utils.TestBwCall) utils.TestBwReply {
		return utils.TestBwReply{Output: []byte(`{"id":"i1","name":"proj","notes":"n","type":2,"folderId":"f"}`)}
	})
	c := NewBwClient()
	item, err := c.GetItemByID("i1")
	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Equal(t, "proj", item.Name)
	assert.Equal(t, "n", item.Notes)
}

func TestRealBwClient_GetItemByName_Found(t *testing.T) {
	stubUtilsBw(t, func(call utils.TestBwCall) utils.TestBwReply {
		if len(call.Args) >= 1 && call.Args[0] == "sync" {
			return utils.TestBwReply{}
		}
		if len(call.Args) >= 2 && call.Args[0] == "list" {
			return utils.TestBwReply{Output: []byte(`[{"id":"i1","name":"proj","notes":"","type":2,"folderId":"f"}]`)}
		}
		if len(call.Args) >= 3 && call.Args[0] == "get" && call.Args[1] == "item" {
			return utils.TestBwReply{Output: []byte(`{"id":"i1","name":"proj","notes":"hello","type":2,"folderId":"f"}`)}
		}
		return utils.TestBwReply{Err: errors.New("unexpected")}
	})
	item, err := NewBwClient().GetItemByName("f", "proj")
	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Equal(t, "hello", item.Notes)
}

func TestRealBwClient_ListItemsError(t *testing.T) {
	stubUtilsBw(t, func(call utils.TestBwCall) utils.TestBwReply {
		return utils.TestBwReply{Output: []byte("Vault is locked."), Err: errors.New("exit 1")}
	})
	_, err := NewBwClient().ListItemsInFolder("f")
	require.Error(t, err)
}

func TestRealBwClient_GetItemByName_Nil(t *testing.T) {
	stubUtilsBw(t, func(call utils.TestBwCall) utils.TestBwReply {
		if len(call.Args) >= 1 && call.Args[0] == "sync" {
			return utils.TestBwReply{}
		}
		if len(call.Args) >= 2 && call.Args[0] == "list" {
			return utils.TestBwReply{Output: []byte(`[]`)}
		}
		return utils.TestBwReply{Err: errors.New("unexpected")}
	})
	c := NewBwClient()
	item, err := c.GetItemByName("f", "missing")
	require.NoError(t, err)
	assert.Nil(t, item)
}

func TestRealBwClient_GetDotenvsFolderID(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	stubUtilsBw(t, func(call utils.TestBwCall) utils.TestBwReply {
		return utils.TestBwReply{Output: []byte(`[{"id":"fid","name":"dotenvs"}]`)}
	})
	id, err := NewBwClient().GetDotenvsFolderID()
	require.NoError(t, err)
	assert.Equal(t, "fid", id)
}

func TestRealBwClient_DotenvsFolderExists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	stubUtilsBw(t, func(call utils.TestBwCall) utils.TestBwReply {
		return utils.TestBwReply{Output: []byte(`[{"id":"fid","name":"dotenvs"}]`)}
	})
	ok, err := NewBwClient().DotenvsFolderExists()
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestRealBwClient_CreateDotenvsFolder(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	stubUtilsBw(t, func(call utils.TestBwCall) utils.TestBwReply {
		return utils.TestBwReply{Output: []byte(`{}`)}
	})
	require.NoError(t, NewBwClient().CreateDotenvsFolder())
}

func TestRealBwClient_CreateAndUpdateNote(t *testing.T) {
	stubUtilsBw(t, func(call utils.TestBwCall) utils.TestBwReply {
		if len(call.Args) >= 3 && call.Args[0] == "get" && call.Args[1] == "template" {
			return utils.TestBwReply{Err: errors.New("no template")}
		}
		if len(call.Args) >= 2 && call.Args[0] == "create" && call.Args[1] == "item" {
			return utils.TestBwReply{Output: []byte(`{}`)}
		}
		if len(call.Args) >= 3 && call.Args[0] == "get" && call.Args[1] == "item" {
			return utils.TestBwReply{Output: []byte(`{"id":"i1","name":"n","notes":"old"}`)}
		}
		if len(call.Args) >= 1 && call.Args[0] == "encode" {
			return utils.TestBwReply{Output: []byte("enc")}
		}
		if len(call.Args) >= 2 && call.Args[0] == "edit" {
			return utils.TestBwReply{}
		}
		return utils.TestBwReply{Err: errors.New("unexpected")}
	})
	c := NewBwClient()
	require.NoError(t, c.CreateNoteItem("f", "n", "notes"))
	require.NoError(t, c.UpdateNoteItem("i1", "new"))
}

func TestRealBwClient_LoginUnlock(t *testing.T) {
	stubUtilsBw(t, func(call utils.TestBwCall) utils.TestBwReply {
		if len(call.Args) >= 1 && call.Args[0] == "login" {
			return utils.TestBwReply{Output: []byte(infraTestSession)}
		}
		if len(call.Args) >= 1 && call.Args[0] == "unlock" {
			return utils.TestBwReply{Output: []byte(infraTestSession)}
		}
		return utils.TestBwReply{Err: errors.New("unexpected")}
	})
	c := NewBwClient()
	require.NoError(t, c.Login("a@b.c", "pw", ""))
	require.NoError(t, c.Unlock("pw"))
}

func TestRealBwClient_LoginUnlockErrors(t *testing.T) {
	stubUtilsBw(t, func(call utils.TestBwCall) utils.TestBwReply {
		return utils.TestBwReply{Output: []byte("nope"), Err: errors.New("exit 1")}
	})
	c := NewBwClient()
	err := c.Login("a@b.c", "pw", "")
	require.Error(t, err)
	var le *LoginError
	assert.ErrorAs(t, err, &le)

	err = c.Unlock("pw")
	require.Error(t, err)
	var ue *UnlockError
	assert.ErrorAs(t, err, &ue)
}
