package pkx

import (
 "context"
 "fmt"
 "io"
 "path/filepath"
 "strings"
)

type universalFamily struct { id string; aliases []string; variants map[string]string }

var curatedFamilies = []universalFamily{
 {"http-client",[]string{"http-client","axios","httpx","requests","guzzle","guzzlehttp/guzzle","reqwest","resty"},map[string]string{"node":"axios","python":"httpx","php":"guzzlehttp/guzzle","rust":"reqwest","go":"github.com/go-resty/resty/v2"}},
 {"structured-logging",[]string{"structured-logging","pino","structlog","monolog","monolog/monolog","tracing","zap"},map[string]string{"node":"pino","python":"structlog","php":"monolog/monolog","rust":"tracing","go":"go.uber.org/zap"}},
 {"schema-validation",[]string{"schema-validation","zod","pydantic","symfony-validator","symfony/validator","validator"},map[string]string{"node":"zod","python":"pydantic","php":"symfony/validator","rust":"validator","go":"github.com/go-playground/validator/v10"}},
}

func packageInstallRequested(args []string) bool {
 for i:=0;i<len(args);i++ {
  v:=args[i]
  if v=="--dir" || v=="--project" || v=="--manager" || v=="--depth" || v=="--only" {i++;continue}
  if strings.HasPrefix(v,"--dir=")||strings.HasPrefix(v,"--project=")||strings.HasPrefix(v,"--manager=")||strings.HasPrefix(v,"--depth=")||strings.HasPrefix(v,"--only=") {continue}
  if !strings.HasPrefix(v,"-") {return true}
 }
 return false
}

func resolveCurated(query, ecosystem string) (string,string,bool) {
 for _,f:=range curatedFamilies {
  for _,a:=range f.aliases {
   if strings.EqualFold(query,a) {
    name,ok:=f.variants[ecosystem];return name,f.id,ok
   }
  }
 }
 return query,"",true // unknown queries are treated as exact native package names
}

func runUniversalInstall(ctx context.Context,args []string,stdout,stderr io.Writer) int {
 opts,packages,_,err:=parseManageArgs("add",args)
 if err!=nil {fmt.Fprintf(stderr,"pkx: %v\n",err);return 2}
 root,err:=rootPath(opts.Dir);if err!=nil {fmt.Fprintf(stderr,"pkx: %v\n",err);return 1}
 projects,err:=Discover(root,4);if err!=nil {fmt.Fprintf(stderr,"pkx: %v\n",err);return 1}
 var candidates []Project
 chosen:=""
 if opts.Project!="" {d,e:=within(root,opts.Project);if e!=nil {fmt.Fprintf(stderr,"pkx: %v\n",e);return 1};chosen=relative(root,d)}
 for _,p:=range projects {
  if chosen!="" && p.Directory!=chosen {continue}
  if opts.Manager!="" && p.Manager!=opts.Manager && p.Ecosystem!=opts.Manager {continue}
  candidates=append(candidates,p)
 }
 if len(candidates)!=1 {fmt.Fprintln(stderr,"pkx: select exactly one project (pkx scan, then --project DIR)");return 1}
 p:=candidates[0]
 var resolved []string
 for _,name:=range packages {
  if err:=validateSpec(name);err!=nil {fmt.Fprintf(stderr,"pkx: %v\n",err);return 2}
  target,family,ok:=resolveCurated(name,p.Ecosystem)
  if !ok {fmt.Fprintf(stderr,"pkx: no reviewed %s equivalent for %s\n",p.Ecosystem,name);return 1}
  if err:=validateSpec(target);err!=nil {fmt.Fprintf(stderr,"pkx: %v\n",err);return 2}
  if family!="" {fmt.Fprintf(stdout,"pkx: %s → %s (%s; functionally similar, not API-compatible)\n",name,target,p.Ecosystem)}
  resolved=append(resolved,target)
 }
 dir:=filepath.Join(root,filepath.FromSlash(p.Directory))
 var step Step
 switch p.Manager {
 case "npm","pnpm","yarn","bun","composer":
  step,err=manageStep(root,p,"add",resolved,nil,opts.Dev)
 case "pip":
  if opts.Dev {err=fmt.Errorf("--dev unsupported for pip");break}
  step=Step{Directory:dir,Manager:"pip",Binary:"python",Args:append([]string{"-m","pip","install"},resolved...)}
 case "uv":
  a:=[]string{"add"};if opts.Dev {a=append(a,"--dev")};step=Step{Directory:dir,Manager:"uv",Binary:"uv",Args:append(a,resolved...)}
 case "poetry":
  a:=[]string{"add"};if opts.Dev {a=append(a,"--group","dev")};step=Step{Directory:dir,Manager:"poetry",Binary:"poetry",Args:append(a,resolved...)}
 case "cargo":
  a:=[]string{"add"};if opts.Dev {a=append(a,"--dev")};step=Step{Directory:dir,Manager:"cargo",Binary:"cargo",Args:append(a,resolved...)}
 case "go":
  if opts.Dev {err=fmt.Errorf("--dev unsupported for Go modules");break}
  step=Step{Directory:dir,Manager:"go",Binary:"go",Args:append([]string{"get"},resolved...)}
 default: err=fmt.Errorf("unsupported manager %s",p.Manager)
 }
 if err!=nil {fmt.Fprintf(stderr,"pkx: %v\n",err);return 1}
 return execute(ctx,root,[]Step{step},opts.DryRun,stdout,stderr)
}
