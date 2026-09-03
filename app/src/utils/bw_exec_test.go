package utils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSessionKey = "P4tHpDULkFR5+NLL1lbfxD43q9NqIS2tmKnG0GMAn/Ft8w4JOipXty4uY4EQ5/gkTXDPGpidXuoC155F65X5sQ=="

func stubBw(t *testing.T, runner bwRunner) {
	t.Helper()
	origRun := runBw
	origLook := lookPath
	runBw = runner
	lookPath = func(file string) (string, error) {
		if file == "bw" {
			return "/mock/bw", nil
		}
		return "", errors.New("not found")
	}
	t.Setenv("NO_COLOR", "1")
	t.Cleanup(func() {
		runBw = origRun
		lookPath = origLook
		_ = os.Unsetenv("BW_SESSION")
	})
}

func stubBwMissing(t *testing.T) {
	t.Helper()
	origLook := lookPath
	lookPath = func(string) (string, error) { return "", errors.New("not found") }
	t.Cleanup(func() { lookPath = origLook })
}

func argsMatch(args []string, want ...string) bool {
	if len(args) < len(want) {
		return false
	}
	for i := range want {
		if args[i] != want[i] {
			return false
		}
	}
	return true
}

func TestGetFolderID_Success(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		assert.Equal(t, "bw", name)
		assert.True(t, argsMatch(args, "list", "folders"))
		return bwResult{Output: []byte(`[{"id":"fid-1","name":"dotenvs"},{"id":"fid-2","name":"other"}]`)}
	})
	id, err := GetFolderID("dotenvs")
	require.NoError(t, err)
	assert.Equal(t, "fid-1", id)
}

func TestGetFolderID_NotFound(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		return bwResult{Output: []byte(`[{"id":"fid-1","name":"other"}]`)}
	})
	_, err := GetFolderID("dotenvs")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "folder not found")
}

func TestGetFolderID_AuthRequired(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		return bwResult{Output: []byte("Vault is locked."), Err: errors.New("exit 1")}
	})
	_, err := GetFolderID("dotenvs")
	assert.ErrorIs(t, err, ErrBitwardenLocked)
}

func TestGetFolderID_InvalidJSON(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		return bwResult{Output: []byte(`[{"id":`)}
	})
	_, err := GetFolderID("dotenvs")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestGetFolderID_BwMissing(t *testing.T) {
	stubBwMissing(t)
	_, err := GetFolderID("dotenvs")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
}

func TestFolderExists(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		return bwResult{Output: []byte(`[{"id":"fid-1","name":"dotenvs"}]`)}
	})
	ok, err := FolderExists("dotenvs")
	require.NoError(t, err)
	assert.True(t, ok)

	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		return bwResult{Output: []byte(`[]`)}
	})
	ok, err = FolderExists("dotenvs")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestCreateFolder_Success(t *testing.T) {
	var sawCreate, sawSync bool
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		if argsMatch(args, "create", "folder") {
			sawCreate = true
			decoded, err := base64.StdEncoding.DecodeString(args[2])
			require.NoError(t, err)
			assert.Contains(t, string(decoded), "dotenvs")
			return bwResult{Output: []byte(`{"id":"new"}`)}
		}
		if argsMatch(args, "sync") {
			sawSync = true
			return bwResult{}
		}
		return bwResult{Err: fmt.Errorf("unexpected %v", args)}
	})
	require.NoError(t, CreateFolder("dotenvs"))
	assert.True(t, sawCreate)
	assert.True(t, sawSync)
}

func TestListItemsInFolder_Success(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		assert.True(t, argsMatch(args, "list", "items", "--folderid", "fid"))
		return bwResult{Output: []byte(`[{"id":"i1","name":"proj"}]`)}
	})
	items, err := ListItemsInFolder("fid")
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "proj", items[0].Name)
}

func TestListItemsInFolder_Auth(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		return bwResult{Output: []byte("You are not logged in."), Err: errors.New("exit 1")}
	})
	_, err := ListItemsInFolder("fid")
	assert.ErrorIs(t, err, ErrBitwardenLocked)
}

func TestBwUnlock_SuccessViaUnlockRaw(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		assert.True(t, argsMatch(args, "unlock", "--raw", "--passwordenv", "BW_PASSWORD"))
		return bwResult{Output: []byte(testSessionKey)}
	})
	ok, msg := BwUnlock("secret")
	assert.True(t, ok)
	assert.Empty(t, msg)
	assert.Equal(t, testSessionKey, os.Getenv("BW_SESSION"))
}

func TestBwUnlock_FallbackPasswordFile(t *testing.T) {
	calls := 0
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		calls++
		if argsMatch(args, "unlock", "--raw", "--passwordenv", "BW_PASSWORD") {
			return bwResult{Output: []byte("fail"), Err: errors.New("exit 1")}
		}
		if argsMatch(args, "unlock", "--raw", "--passwordfile") {
			assert.True(t, opts.SeparateStreams)
			return bwResult{Output: []byte(testSessionKey)}
		}
		return bwResult{Err: fmt.Errorf("unexpected %v", args)}
	})
	ok, msg := BwUnlock("secret")
	assert.True(t, ok, msg)
	assert.Empty(t, msg)
	assert.GreaterOrEqual(t, calls, 2)
}

func TestCheckBwCommand_LookPath(t *testing.T) {
	origLook := lookPath
	t.Cleanup(func() { lookPath = origLook })

	lookPath = func(file string) (string, error) { return "/usr/bin/bw", nil }
	ok, path := CheckBwCommand()
	assert.True(t, ok)
	assert.Equal(t, "/usr/bin/bw", path)

	lookPath = func(string) (string, error) { return "", errors.New("not found") }
	ok, path = CheckBwCommand()
	assert.False(t, ok)
	assert.Empty(t, path)
}

func TestBwLogin_SuccessRawSession(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		assert.True(t, argsMatch(args, "login", "a@b.c", "pw", "--raw"))
		return bwResult{Output: []byte(testSessionKey)}
	})
	ok, msg := BwLogin("a@b.c", "pw", "")
	assert.True(t, ok, msg)
	assert.Empty(t, msg)
	assert.Equal(t, testSessionKey, os.Getenv("BW_SESSION"))
}

func TestBwLogin_AlreadyLoggedInUnlocks(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		if argsMatch(args, "login") {
			return bwResult{Output: []byte("You are already logged in."), Err: errors.New("exit 1")}
		}
		if argsMatch(args, "unlock", "--raw", "--passwordenv", "BW_PASSWORD") {
			return bwResult{Output: []byte(testSessionKey)}
		}
		return bwResult{Err: fmt.Errorf("unexpected %v", args)}
	})
	ok, msg := BwLogin("a@b.c", "pw", "")
	assert.True(t, ok, msg)
}

func TestBwLogin_ConfigServer(t *testing.T) {
	var sawConfig bool
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		if argsMatch(args, "config", "server") && len(args) == 2 {
			return bwResult{Output: []byte("https://old.example")}
		}
		if argsMatch(args, "logout") {
			return bwResult{}
		}
		if argsMatch(args, "config", "server", "https://new.example") {
			sawConfig = true
			return bwResult{}
		}
		if argsMatch(args, "login") {
			return bwResult{Output: []byte(testSessionKey)}
		}
		return bwResult{Err: fmt.Errorf("unexpected %v", args)}
	})
	ok, msg := BwLogin("a@b.c", "pw", "https://new.example")
	assert.True(t, ok, msg)
	assert.True(t, sawConfig)
}

func TestBwLogin_Failure(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		return bwResult{Output: []byte("bad credentials"), Err: errors.New("exit 1")}
	})
	ok, msg := BwLogin("a@b.c", "pw", "")
	assert.False(t, ok)
	assert.Contains(t, msg, "bad credentials")
}

func TestGetItemByID_Success(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		assert.True(t, argsMatch(args, "get", "item", "iid"))
		return bwResult{Output: []byte(`{"id":"iid","name":"proj","notes":"{}","type":2,"folderId":"f"}`)}
	})
	item, err := GetItemByID("iid")
	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Equal(t, "proj", item.Name)
}

func TestGetItemByID_Auth(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		return bwResult{Output: []byte("Master password required"), Err: errors.New("exit 1")}
	})
	_, err := GetItemByID("iid")
	assert.ErrorIs(t, err, ErrBitwardenLocked)
}

func TestGetItemByName_Found(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		if argsMatch(args, "sync") {
			return bwResult{}
		}
		if argsMatch(args, "list", "items") {
			return bwResult{Output: []byte(`[{"id":"iid","name":"proj","notes":"","type":2,"folderId":"f"}]`)}
		}
		if argsMatch(args, "get", "item", "iid") {
			return bwResult{Output: []byte(`{"id":"iid","name":"proj","notes":"{\"x\":1}","type":2,"folderId":"f"}`)}
		}
		return bwResult{Err: fmt.Errorf("unexpected %v", args)}
	})
	item, err := GetItemByName("f", "proj")
	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Equal(t, "iid", item.ID)
}

func TestGetItemByName_NotFound(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		if argsMatch(args, "sync") {
			return bwResult{}
		}
		if argsMatch(args, "list", "items") {
			return bwResult{Output: []byte(`[]`)}
		}
		return bwResult{Err: fmt.Errorf("unexpected %v", args)}
	})
	item, err := GetItemByName("f", "missing")
	require.NoError(t, err)
	assert.Nil(t, item)
}

func TestCreateNoteItem_EncodePath(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		if argsMatch(args, "get", "template", "item") {
			return bwResult{Output: []byte(`{"type":1,"name":"","notes":""}`)}
		}
		if argsMatch(args, "encode") {
			assert.NotEmpty(t, opts.Stdin)
			assert.True(t, opts.StdoutOnly)
			return bwResult{Output: []byte("encoded-payload")}
		}
		if argsMatch(args, "create", "item", "encoded-payload") {
			return bwResult{Output: []byte(`{"id":"new"}`)}
		}
		return bwResult{Err: fmt.Errorf("unexpected %v", args)}
	})
	require.NoError(t, CreateNoteItem("f", "n", `{"lines":["A=1"]}`))
}

func TestCreateNoteItem_DirectFallback(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		if argsMatch(args, "get", "template", "item") {
			return bwResult{Err: errors.New("no template"), Output: []byte("err")}
		}
		if argsMatch(args, "create", "item") && len(args) == 2 {
			assert.NotEmpty(t, opts.Stdin)
			return bwResult{Output: []byte(`{"id":"new"}`)}
		}
		return bwResult{Err: fmt.Errorf("unexpected %v", args)}
	})
	require.NoError(t, CreateNoteItem("f", "n", "notes"))
}

func TestCreateNoteItem_BothFail(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		if argsMatch(args, "get", "template", "item") {
			return bwResult{Err: errors.New("no template")}
		}
		if argsMatch(args, "create", "item") {
			return bwResult{Output: []byte("boom"), Err: errors.New("exit 1")}
		}
		return bwResult{Err: fmt.Errorf("unexpected %v", args)}
	})
	err := CreateNoteItem("f", "n", "notes")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create item")
}

func TestUpdateNoteItem_EncodeEdit(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		if argsMatch(args, "get", "item", "iid") {
			return bwResult{Output: []byte(`{"id":"iid","name":"n","notes":"old"}`)}
		}
		if argsMatch(args, "encode") {
			return bwResult{Output: []byte("enc")}
		}
		if argsMatch(args, "edit", "item", "iid", "enc") {
			return bwResult{}
		}
		return bwResult{Err: fmt.Errorf("unexpected %v", args)}
	})
	require.NoError(t, UpdateNoteItem("iid", "new-notes"))
}

func TestUpdateNoteItem_DirectEditFallback(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		if argsMatch(args, "get", "item", "iid") {
			return bwResult{Output: []byte(`{"id":"iid","name":"n","notes":"old"}`)}
		}
		if argsMatch(args, "encode") {
			return bwResult{Err: errors.New("encode fail")}
		}
		if argsMatch(args, "edit", "item", "iid") && len(args) == 3 {
			var got map[string]interface{}
			require.NoError(t, json.Unmarshal(opts.Stdin, &got))
			assert.Equal(t, "new", got["notes"])
			return bwResult{}
		}
		return bwResult{Err: fmt.Errorf("unexpected %v", args)}
	})
	require.NoError(t, UpdateNoteItem("iid", "new"))
}

func TestUpdateNoteItem_GetFails(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		return bwResult{Output: []byte("missing"), Err: errors.New("exit 1")}
	})
	err := UpdateNoteItem("iid", "n")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get item")
}

func TestResolveConfiguredFolderName_Default(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	assert.Equal(t, "dotenvs", resolveConfiguredFolderName())
}

func TestDefaultRunBw_CombinedAndStdin(t *testing.T) {
	res := defaultRunBw("echo", []string{"hello-bw"}, bwRunOptions{})
	require.NoError(t, res.Err)
	assert.Contains(t, string(res.Output), "hello-bw")

	res = defaultRunBw("cat", nil, bwRunOptions{Stdin: []byte("piped\n")})
	require.NoError(t, res.Err)
	assert.Contains(t, string(res.Output), "piped")

	res = defaultRunBw("echo", []string{"only"}, bwRunOptions{StdoutOnly: true})
	require.NoError(t, res.Err)
	assert.Contains(t, string(res.Output), "only")

	res = defaultRunBw("echo", []string{"sep"}, bwRunOptions{SeparateStreams: true})
	require.NoError(t, res.Err)
	assert.Contains(t, string(res.Output), "sep")

	res = defaultRunBw("false", nil, bwRunOptions{})
	assert.Error(t, res.Err)
}

func TestInstallBwExecTestHook_RoundTrip(t *testing.T) {
	restore := InstallBwExecTestHook(
		func(file string) (string, error) { return "/hook/bw", nil },
		func(call TestBwCall) TestBwReply {
			return TestBwReply{Output: []byte("ok")}
		},
	)
	t.Cleanup(restore)

	ok, path := CheckBwCommand()
	assert.True(t, ok)
	assert.Equal(t, "/hook/bw", path)

	res := runBw("bw", []string{"version"}, bwRunOptions{})
	assert.Equal(t, "ok", string(res.Output))
}

func TestGetDotenvsFolderID_AndCreate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		if argsMatch(args, "list", "folders") {
			return bwResult{Output: []byte(`[{"id":"fid","name":"dotenvs"}]`)}
		}
		if argsMatch(args, "create", "folder") || argsMatch(args, "sync") {
			return bwResult{Output: []byte(`{}`)}
		}
		return bwResult{Err: fmt.Errorf("unexpected %v", args)}
	})
	id, err := GetDotenvsFolderID()
	require.NoError(t, err)
	assert.Equal(t, "fid", id)
	require.NoError(t, CreateDotenvsFolder())
	ok, err := DotenvsFolderExists()
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestGetFolderID_EmptyNameAndNonJSON(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		return bwResult{Output: []byte(`[{"id":"fid","name":"dotenvs"}]`)}
	})
	id, err := GetFolderID("")
	require.NoError(t, err)
	assert.Equal(t, "fid", id)

	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		return bwResult{Output: []byte("not-json-output")}
	})
	_, err = GetFolderID("dotenvs")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not JSON")

	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		return bwResult{Output: []byte("   ")}
	})
	_, err = GetFolderID("dotenvs")
	require.Error(t, err)
}

func TestListItemsInFolder_EmptyAndNonJSON(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		return bwResult{Output: []byte("   ")}
	})
	_, err := ListItemsInFolder("f")
	require.Error(t, err)

	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		return bwResult{Output: []byte("Master password on stdout")}
	})
	_, err = ListItemsInFolder("f")
	assert.ErrorIs(t, err, ErrBitwardenLocked)

	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		return bwResult{Output: []byte("plain text")}
	})
	_, err = ListItemsInFolder("f")
	require.Error(t, err)
}

func TestBwUnlock_EmptyPassword(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		return bwResult{}
	})
	ok, msg := BwUnlock("")
	assert.False(t, ok)
	assert.Contains(t, msg, "empty")
}

func TestBwUnlock_FallbackFails(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		if argsMatch(args, "unlock", "--raw", "--passwordenv") {
			return bwResult{Output: []byte("fail"), Err: errors.New("exit 1")}
		}
		if argsMatch(args, "unlock", "--raw", "--passwordfile") {
			return bwResult{Output: []byte("still fail"), Stderr: []byte("bad"), Err: errors.New("exit 1")}
		}
		return bwResult{Err: fmt.Errorf("unexpected %v", args)}
	})
	ok, msg := BwUnlock("secret")
	assert.False(t, ok)
	assert.NotEmpty(t, msg)
}

func TestCreateFolder_EmptyNameAndError(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		if argsMatch(args, "create", "folder") {
			return bwResult{Output: []byte("nope"), Err: errors.New("exit 1")}
		}
		return bwResult{}
	})
	err := CreateFolder("")
	require.Error(t, err)
}

func TestGetItemByName_SyncWarningAndAuth(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		if argsMatch(args, "sync") {
			return bwResult{Output: []byte("network error"), Err: errors.New("exit 1")}
		}
		if argsMatch(args, "list", "items") {
			return bwResult{Output: []byte("Vault is locked."), Err: errors.New("exit 1")}
		}
		return bwResult{Err: fmt.Errorf("unexpected %v", args)}
	})
	_, err := GetItemByName("f", "x")
	assert.ErrorIs(t, err, ErrBitwardenLocked)
}

func TestGetItemByID_EmptyAndNonJSON(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		return bwResult{Output: []byte("  ")}
	})
	_, err := GetItemByID("i")
	require.Error(t, err)

	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		return bwResult{Output: []byte("not-json")}
	})
	_, err = GetItemByID("i")
	require.Error(t, err)
}

func TestCreateNoteItem_EncodeCreateFails(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		if argsMatch(args, "get", "template", "item") {
			return bwResult{Output: []byte(`{"type":1}`)}
		}
		if argsMatch(args, "encode") {
			return bwResult{Output: []byte("enc")}
		}
		if argsMatch(args, "create", "item", "enc") {
			return bwResult{Output: []byte("fail"), Err: errors.New("exit 1")}
		}
		return bwResult{Err: fmt.Errorf("unexpected %v", args)}
	})
	err := CreateNoteItem("f", "n", "notes")
	require.Error(t, err)
}

func TestUpdateNoteItem_EditFails(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		if argsMatch(args, "get", "item") {
			return bwResult{Output: []byte(`{"id":"i","notes":"o"}`)}
		}
		if argsMatch(args, "encode") {
			return bwResult{Output: []byte("enc")}
		}
		if argsMatch(args, "edit", "item") {
			return bwResult{Output: []byte("fail"), Err: errors.New("exit 1")}
		}
		return bwResult{Err: fmt.Errorf("unexpected %v", args)}
	})
	err := UpdateNoteItem("i", "n")
	require.Error(t, err)
}

func TestBwLogin_LoggedInMessageUnlocks(t *testing.T) {
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		if argsMatch(args, "login") {
			return bwResult{Output: []byte("You are logged in!")}
		}
		if argsMatch(args, "unlock") {
			return bwResult{Output: []byte(testSessionKey)}
		}
		return bwResult{Err: fmt.Errorf("unexpected %v", args)}
	})
	ok, msg := BwLogin("a@b.c", "pw", "")
	assert.True(t, ok, msg)
}

func TestBwLogin_ConfigLogoutRequired(t *testing.T) {
	configCalls := 0
	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		if argsMatch(args, "config", "server") && len(args) == 2 {
			return bwResult{Output: []byte("")}
		}
		if argsMatch(args, "config", "server") && len(args) == 3 {
			configCalls++
			if configCalls == 1 {
				return bwResult{Output: []byte("Logout required"), Err: errors.New("exit 1")}
			}
			return bwResult{}
		}
		if argsMatch(args, "logout") {
			return bwResult{}
		}
		if argsMatch(args, "login") {
			return bwResult{Output: []byte(testSessionKey)}
		}
		return bwResult{Err: fmt.Errorf("unexpected %v", args)}
	})
	ok, msg := BwLogin("a@b.c", "pw", "https://new.example")
	assert.True(t, ok, msg)
}

func TestBwLogin_MissingBw(t *testing.T) {
	stubBwMissing(t)
	ok, msg := BwLogin("a", "b", "")
	assert.False(t, ok)
	assert.Contains(t, msg, "not installed")
}

func TestUnlockRaw_EmptyAndNoSession(t *testing.T) {
	_, err := unlockRaw("")
	require.Error(t, err)

	stubBw(t, func(name string, args []string, opts bwRunOptions) bwResult {
		return bwResult{Output: []byte("no-key-here")}
	})
	_, err = unlockRaw("pw")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no session")
}
