package pkx

import "testing"

func TestCuratedResolutionAcrossLanguages(t *testing.T) {
 cases:=[]struct{query,ecosystem,want string}{
  {"axios","node","axios"},
  {"axios","python","httpx"},
  {"axios","php","guzzlehttp/guzzle"},
  {"axios","rust","reqwest"},
  {"axios","go","github.com/go-resty/resty/v2"},
  {"zod","python","pydantic"},
  {"pino","php","monolog/monolog"},
  {"PINO","rust","tracing"},
 }
 for _,tc:=range cases{
  got,family,ok:=resolveCurated(tc.query,tc.ecosystem)
  if !ok || family=="" || got!=tc.want {t.Errorf("%s -> %s: got %q family=%q ok=%v, want %q",tc.query,tc.ecosystem,got,family,ok,tc.want)}
 }
}
func TestUnknownPackageIsNotGuessed(t *testing.T) {
 got,family,ok:=resolveCurated("my-exact-package","python")
 if !ok || family!="" || got!="my-exact-package" {t.Fatalf("unexpected resolution: %q %q %v",got,family,ok)}
}
func TestInstallCommandDetection(t *testing.T) {
 cases:=[]struct{args []string;want bool}{
  {[]string{"axios"},true},
  {[]string{"--dir","."},false},
  {[]string{"--dir",".","axios"},true},
  {[]string{"--dir=.","--dry-run","axios"},true},
  {[]string{"--only","npm"},false},
  {[]string{"--dry-run"},false},
 }
 for _,tc:=range cases {if got:=packageInstallRequested(tc.args);got!=tc.want {t.Errorf("%q: got %v want %v",tc.args,got,tc.want)}}
}
