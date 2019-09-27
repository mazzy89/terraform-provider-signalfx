package signalfx

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/stretchr/testify/assert"
)

var OldSystemConfigPath = SystemConfigPath
var OldHomeConfigPath = HomeConfigPath

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"signalfx": testAccProvider,
	}
}

func resetGlobals() {
	SystemConfigPath = OldSystemConfigPath
	HomeConfigPath = OldHomeConfigPath
}

func createTempConfigFile(content string, name string) (*os.File, error) {
	tmpfile, err := ioutil.TempFile(os.TempDir(), name)
	if err != nil {
		return nil, fmt.Errorf("Error creating temporary test file. err: %s", err.Error())
	}

	_, err = tmpfile.WriteString(content)
	if err != nil {
		os.Remove(tmpfile.Name())
		return nil, fmt.Errorf("Error writing to temporary test file. err: %s", err.Error())
	}

	return tmpfile, nil
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatal(err.Error())
	}
}

func TestProviderConfigureFromNothing(t *testing.T) {
	defer resetGlobals()

	old := os.Getenv("SFX_AUTH_TOKEN")
	defer os.Setenv("SFX_AUTH_TOKEN", old)
	os.Unsetenv("SFX_AUTH_TOKEN")

	SystemConfigPath = "filedoesnotexist"
	HomeConfigPath = "filedoesnotexist"
	raw := make(map[string]interface{})

	rp := Provider()
	err := rp.Configure(terraform.NewResourceConfigRaw(raw))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "auth_token: required field is not set")
}

func TestProviderConfigureFromTerraform(t *testing.T) {
	defer resetGlobals()
	tmpfileSystem, err := createTempConfigFile(`{"useless_config":"foo","auth_token":"ZZZ"}`, "signalfx.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileSystem.Name())
	SystemConfigPath = tmpfileSystem.Name()
	tmpfileHome, err := createTempConfigFile(`{"auth_token":"WWW"}`, "signalfx.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileHome.Name())
	old := os.Getenv("SFX_AUTH_TOKEN")
	os.Setenv("SFX_AUTH_TOKEN", "YYY")
	defer os.Setenv("SFX_AUTH_TOKEN", old)
	raw := map[string]interface{}{
		"auth_token": "XXX",
	}

	rp := Provider()
	err = rp.Configure(terraform.NewResourceConfigRaw(raw))
	meta := rp.(*schema.Provider).Meta()
	if meta == nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	}
	configuration := meta.(*signalfxConfig)
	assert.Equal(t, "XXX", configuration.AuthToken)
}

func TestProviderConfigureFromTerraformOnly(t *testing.T) {
	defer resetGlobals()
	SystemConfigPath = "filedoesnotexist"
	HomeConfigPath = "filedoesnotexist"
	raw := map[string]interface{}{
		"auth_token": "XXX",
	}

	rp := Provider()
	err := rp.Configure(terraform.NewResourceConfigRaw(raw))
	meta := rp.(*schema.Provider).Meta()
	if meta == nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	}
	configuration := meta.(*signalfxConfig)
	assert.Equal(t, "XXX", configuration.AuthToken)
}

func TestProviderConfigureFromEnvironment(t *testing.T) {
	defer resetGlobals()
	tmpfileSystem, err := createTempConfigFile(`{"useless_config":"foo","auth_token":"ZZZ"}`, "signalfx.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileSystem.Name())
	SystemConfigPath = tmpfileSystem.Name()
	tmpfileHome, err := createTempConfigFile(`{"auth_token":"WWW"}`, "signalfx.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileHome.Name())
	old := os.Getenv("SFX_AUTH_TOKEN")
	os.Setenv("SFX_AUTH_TOKEN", "YYY")
	defer os.Setenv("SFX_AUTH_TOKEN", old)
	raw := make(map[string]interface{})

	rp := Provider()
	err = rp.Configure(terraform.NewResourceConfigRaw(raw))
	meta := rp.(*schema.Provider).Meta()
	if meta == nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	}
	configuration := meta.(*signalfxConfig)
	assert.Equal(t, "YYY", configuration.AuthToken)
}

func TestProviderConfigureFromEnvironmentOnly(t *testing.T) {
	defer resetGlobals()
	SystemConfigPath = "filedoesnotexist"
	HomeConfigPath = "filedoesnotexist"
	old := os.Getenv("SFX_AUTH_TOKEN")
	os.Setenv("SFX_AUTH_TOKEN", "YYY")
	defer os.Setenv("SFX_AUTH_TOKEN", old)
	raw := make(map[string]interface{})

	rp := Provider()
	err := rp.Configure(terraform.NewResourceConfigRaw(raw))
	meta := rp.(*schema.Provider).Meta()
	if meta == nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	}
	configuration := meta.(*signalfxConfig)
	assert.Equal(t, "YYY", configuration.AuthToken)
}

func TestSignalFxConfigureFromHomeFile(t *testing.T) {
	defer resetGlobals()
	tmpfileSystem, err := createTempConfigFile(`{"useless_config":"foo","auth_token":"ZZZ"}`, "signalfx.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileSystem.Name())

	old := os.Getenv("SFX_AUTH_TOKEN")
	defer os.Setenv("SFX_AUTH_TOKEN", old)
	os.Unsetenv("SFX_AUTH_TOKEN")

	SystemConfigPath = tmpfileSystem.Name()
	tmpfileHome, err := createTempConfigFile(`{"auth_token":"WWW"}`, "signalfx.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileHome.Name())
	HomeConfigPath = tmpfileHome.Name()
	raw := make(map[string]interface{})

	rp := Provider()
	err = rp.Configure(terraform.NewResourceConfigRaw(raw))
	meta := rp.(*schema.Provider).Meta()
	if meta == nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	}
	configuration := meta.(*signalfxConfig)
	assert.Equal(t, "WWW", configuration.AuthToken)
}

func TestSignalFxConfigureFromNetrcFile(t *testing.T) {
	defer resetGlobals()
	tmpfileSystem, err := createTempConfigFile(`{"useless_config":"foo","auth_token":"ZZZ"}`, "signalfx.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileSystem.Name())
	SystemConfigPath = tmpfileSystem.Name()
	tmpfileHome, err := createTempConfigFile(`machine api.signalfx.com login auth_login password WWW`, ".netrc")
	if err != nil {
		t.Fatal(err.Error())
	}
	old := os.Getenv("SFX_AUTH_TOKEN")
	defer os.Setenv("SFX_AUTH_TOKEN", old)
	os.Unsetenv("SFX_AUTH_TOKEN")

	defer os.Remove(tmpfileHome.Name())
	os.Setenv("NETRC", tmpfileHome.Name())
	defer os.Unsetenv("NETRC")
	raw := make(map[string]interface{})

	rp := Provider()
	err = rp.Configure(terraform.NewResourceConfigRaw(raw))
	meta := rp.(*schema.Provider).Meta()
	if meta == nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	}
	configuration := meta.(*signalfxConfig)
	assert.Equal(t, "WWW", configuration.AuthToken)
}

func TestSignalFxConfigureFromHomeFileOnly(t *testing.T) {
	defer resetGlobals()
	SystemConfigPath = "filedoesnotexist"
	tmpfileHome, err := createTempConfigFile(`{"auth_token":"WWW"}`, "signalfx.conf")
	if err != nil {
		t.Fatal(err.Error())
	}

	old := os.Getenv("SFX_AUTH_TOKEN")
	defer os.Setenv("SFX_AUTH_TOKEN", old)
	os.Unsetenv("SFX_AUTH_TOKEN")

	defer os.Remove(tmpfileHome.Name())
	HomeConfigPath = tmpfileHome.Name()
	raw := make(map[string]interface{})

	rp := Provider()
	err = rp.Configure(terraform.NewResourceConfigRaw(raw))
	meta := rp.(*schema.Provider).Meta()
	if meta == nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	}
	configuration := meta.(*signalfxConfig)
	assert.Equal(t, "WWW", configuration.AuthToken)
}

func TestSignalFxConfigureFromSystemFileOnly(t *testing.T) {
	defer resetGlobals()

	old := os.Getenv("SFX_AUTH_TOKEN")
	defer os.Setenv("SFX_AUTH_TOKEN", old)
	os.Unsetenv("SFX_AUTH_TOKEN")

	tmpfileSystem, err := createTempConfigFile(`{"useless_config":"foo","auth_token":"ZZZ"}`, "signalfx.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfileSystem.Name())
	SystemConfigPath = tmpfileSystem.Name()
	HomeConfigPath = "filedoesnotexist"
	raw := make(map[string]interface{})

	rp := Provider()
	err = rp.Configure(terraform.NewResourceConfigRaw(raw))
	meta := rp.(*schema.Provider).Meta()
	if meta == nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	}
	configuration := meta.(*signalfxConfig)
	assert.Equal(t, "ZZZ", configuration.AuthToken)
}

func TestReadConfigFileFileNotFound(t *testing.T) {
	SystemConfigPath = "filedoesnotexist"
	HomeConfigPath = "filedoesnotexist"
	defer resetGlobals()
	config := signalfxConfig{}
	err := readConfigFile("foo.conf", &config)
	assert.Contains(t, err.Error(), "Failed to open config file")
}

func TestReadConfigFileParseError(t *testing.T) {
	config := signalfxConfig{}
	tmpfile, err := createTempConfigFile(`{"auth_tok`, "signalfx.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfile.Name())

	err = readConfigFile(tmpfile.Name(), &config)
	assert.Contains(t, err.Error(), "Failed to parse config file")
}

func TestReadConfigFileSuccess(t *testing.T) {
	config := signalfxConfig{}
	tmpfile, err := createTempConfigFile(`{"useless_config":"foo","auth_token":"XXX"}`, "signalfx.conf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.Remove(tmpfile.Name())

	err = readConfigFile(tmpfile.Name(), &config)
	assert.Nil(t, err)
	assert.Equal(t, "XXX", config.AuthToken)
}
