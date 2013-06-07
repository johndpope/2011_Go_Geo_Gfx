package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"
)

var (
	compiled = map[string]bool {}
	timps = []string { "tshared", "tserver", "thost", "tclient" }
	timp = timps[1:]
	stimp = timps[0]
	timpj = strings.Join(timp, ",")
)

func compileGoFile (filePath string, isDebug bool) {
	var goFile, err = os.Open(filePath)
	var bufReader *bufio.Reader
	var line, tmp, tn string
	var inImps = false
	var imps, args []string
	var cmdOut []byte
	var pos int
	if err != nil {
		panic(err)
	}
	bufReader = bufio.NewReader(goFile)
	for {
		if line, err = bufReader.ReadString('\n'); err != nil {
			break
		}
		if strings.HasPrefix(line, "import (") {
			inImps = true
		} else if inImps && strings.HasPrefix(line, ")") {
			inImps = false
			break
		} else if inImps {
			for _, tn = range timps {
				if pos = strings.Index(line, "\"" + tn + "/"); pos >= 0 {
					tmp = line[pos + 1:]
					if pos = strings.Index(tmp, "\""); pos >= 0 {
						tmp = tmp[0:pos]
					}
					imps = append(imps, tmp)
					break
				}
			}
		}		 
	}
	for _, tn = range imps {
		if !compiled[tn] {
			compileGoFile(path.Join("_src", tn + ".go"), isDebug)
			compiled[tn] = true
		}
	}
	args = []string { "-o", path.Join("_tmp", filePath[5:len(filePath) - 3] + ".6"), "-I", "_tmp", filePath }
	if !isDebug {
		args = append(args, filePath)
		args[len(args) - 2] = "_tmp"
		args[len(args) - 3] = "-I"
		args[len(args) - 4] = "-B"
	}
	cmdOut, err = exec.Command("6g", args...).Output()
	if len(cmdOut) > 0 {
		fmt.Printf("%s\n", cmdOut)
	}
	if err != nil {
		os.Exit(1)
		return
	}
}

func ensureFolderHierarchy (srcDir *os.File, targetDirPath string) {
	var fileInfos, err = srcDir.Readdir(0)
	var srcSubDir, targetSubDir *os.File
	var srcDirPath, subDirName, subDirPath string
	if err != nil {
		panic(err)
	}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			srcDirPath = srcDir.Name()
			subDirName = fileInfo.Name()
			if srcSubDir, err = os.Open(path.Join(srcDirPath, subDirName)); err != nil {
				panic(err)
			}
			subDirPath = path.Join(targetDirPath, subDirName)
			if targetSubDir, err = os.Open(subDirPath); err != nil {
				if err = os.Mkdir(subDirPath, os.ModeDir | os.ModePerm); err != nil {
					panic(err)
				}
			} else {
				targetSubDir.Close()
			}
			ensureFolderHierarchy(srcSubDir, subDirPath)
			srcSubDir.Close()
		}
	}
}

func inSlice (slice []string, val string) bool {
	for _, v := range(slice) {
		if v == val {
			return true
		}
	}
	return false
}

func main () {
	var startTime = time.Now()
	var flagBuild = flag.String("b", "", "build: " + timpj + "... multiples joined with commas, empty=all")
	var flagRun = flag.String("r", "", "run: " + timpj + "... multiples joined with commas, empty=all")
	var flagCmd = flag.String("c", "", "commands to build, multiples joined with commas, empty=all")
	var flagDebug = flag.Bool("d", false, "debug: true to add, false to skip debug symbols")
	var flagWait = flag.Bool("w", false, "wait: true to wait for command(s) launched with -run")
	var flagStFile = flag.String("stf", "", "ST file: current Sublime Text file from which to build")
	var cwd, err = os.Getwd()
	var glslSrc, glslTmp = "package glsl\n\nvar VShaders = map[string]string {}\nvar FShaders = map[string]string {}\n\nfunc init () {\n", ""
	var targets, srcDirPath, fileName, cmdName, lflag, exePath, exeFilePath, exeDirPath, tmpSrc string
	var srcDir *os.File
	var exe *exec.Cmd
	var exeArgs []string
	var exes []*exec.Cmd
	var fileInfo os.FileInfo
	var fileInfos []os.FileInfo
	var cmds, tgts []string
	var cmdOut []byte
	var isFShader, isVShader bool
	var pos1, pos2 int
	if err != nil {
		panic(err)
	}
	runtime.GOMAXPROCS(16)
	flag.Parse()
	for i, arg := range os.Args {
		if (i > 0) && !strings.HasPrefix(arg, "-") {
			if strings.Contains(arg, "=") {
				exeArgs = append(exeArgs, "-" + arg)
			} else {
				exeArgs = append(exeArgs, arg)
			}
		}
	}
	if len(*flagCmd) > 0 {
		cmds = strings.Split(*flagCmd, ",")
	}
	if len(*flagBuild) > 0 {
		targets = *flagBuild
	} else if len(*flagRun) > 0 {
		targets = *flagRun
	}
	if *flagDebug {
		lflag = "-e"
	} else {
		lflag = "-s"
	}
	if len(*flagStFile) > 0 {
		/* /home/roxor/terra/_src/tclient/tclient.go */
		targets = *flagStFile
		targets = targets[strings.Index(targets, "/_src/") + 6:]
		cmds = []string { targets[strings.LastIndex(targets, "/") + 1:] }
		cmds[0] = cmds[0][:len(cmds[0]) - 3]
		targets = targets[:strings.LastIndex(targets, "/")]
	}
	if len(targets) == 0 {
		targets = timpj
	}
	targets = stimp + "," + targets
	tgts = strings.Split(targets, ",")
	for _, targetName := range tgts {
		srcDirPath = path.Join("_src", targetName)
		if srcDir, err = os.Open(srcDirPath); err != nil {
			panic(err)
		}
		ensureFolderHierarchy(srcDir, path.Join("_tmp", targetName))
		srcDir.Close()
	}
	for _, targetName := range tgts {
		if targetName != stimp {
			if targetName == "tclient" {
				srcDirPath = path.Join("_src", targetName, "glsl")
				if srcDir, err = os.Open(srcDirPath); err == nil {
					if fileInfos, err = srcDir.Readdir(0); err == nil {
						srcDir.Close()
						for _, fileInfo = range fileInfos {
							fileName = fileInfo.Name()
							if isFShader = strings.HasSuffix(fileName, ".fs"); isFShader {
								glslTmp = "F"
							}
							if isVShader = strings.HasSuffix(fileName, ".vs"); isVShader {
								glslTmp = "V"
							}
							if isFShader || isVShader {
								glslSrc += fmt.Sprintf("\t%sShaders[\"%s\"] = ", glslTmp, fileName[:len(fileName) - 3])
								glslTmp = "\"\""
								if srcDir, err = os.Open(path.Join(srcDirPath, fileName)); err == nil {
									if cmdOut, err = ioutil.ReadAll(srcDir); err == nil {
										tmpSrc = string(cmdOut)
										if !*flagDebug {
											for {
												if pos1 = strings.Index(tmpSrc, "/*"); pos1 < 0 { break }
												if pos2 = strings.Index(tmpSrc, "*/"); pos2 < pos1 { break }
												tmpSrc = tmpSrc[0:pos1] + tmpSrc[pos2 + 2:]
											}
										}
										glslTmp = fmt.Sprintf("%#v", tmpSrc)
									}
									srcDir.Close()
								}
								glslSrc += (glslTmp + "\n")
							}
						}
						ioutil.WriteFile(path.Join("_src", targetName, "gfx", "glsl.go"), []byte(glslSrc + "}\n"), os.ModePerm)
					} else {
						srcDir.Close()
					}
				}
			}
			srcDirPath = path.Join("_src", targetName)
			if srcDir, err = os.Open(srcDirPath); err != nil {
				panic(err)
			}
			if fileInfos, err = srcDir.Readdir(0); err != nil {
				srcDir.Close()
				panic(err)
			}
			srcDir.Close()
			for _, fileInfo = range fileInfos {
				fileName = fileInfo.Name()
				if strings.HasSuffix(fileName, ".go") {
					cmdName = fileName[:len(fileName) - 3]
					if (len(cmds) == 0) || inSlice(cmds, cmdName) {
						compileGoFile(path.Join(srcDirPath, fileName), *flagDebug)
						exePath = path.Join(targetName, cmdName + ".exe")
						cmdOut, err = exec.Command("6l", "-o", exePath, "-L", "_tmp", lflag, path.Join("_tmp", targetName, cmdName + ".6")).Output()
						if len(cmdOut) > 0 {
							fmt.Printf("%s\n", cmdOut)
						}
						if err != nil {
							os.Exit(1)
							return
						}
						if len(*flagRun) > 0 {
							exeDirPath = path.Join(cwd, targetName)
							exeFilePath = path.Join(cwd, exePath)
							exe = exec.Command(exeFilePath, exeArgs...)
							exe.Dir = exeDirPath
							exe.Stderr = os.Stderr
							exe.Stdin = os.Stdin
							exe.Stdout = os.Stdout
							exes = append(exes, exe)
						}
					}
				}
			}
		}
	}
	fmt.Printf("Built in %v\n", time.Now().Sub(startTime))
	if len(exes) > 0 {
		for _, exe = range exes {
			if *flagWait {
				err = exe.Run()
			} else {
				err = exe.Start()
			}
			if (err != nil) {
				panic(err)
			}
		}
	}
}
