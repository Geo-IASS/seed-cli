/*
Seed implements a command line interface library to build and run
docker images defined by a seed.manifest.json file.
usage is as folllows:
	seed build [OPTIONS]
		Options:
		-d, -directory	The directory containing the seed spec and Dockerfile
										(default is current directory)

	seed init [OPTIONS]
		Options:
		-d, -directory	The directory to create example seed.manifest.json within
										(default is current directory)

	seed list [OPTIONS]
		Not yet implemented

	seed publish [OPTIONS]
		Not yet implemented

	seed run [OPTIONS]
		Options:
		-i, -inputs  The input data. May be multiple -id flags defined
										(seedfile: Job.Interface.Inputs.Files)
		-in, -imageName The name of the Docker image to run (overrides image name
										pulled from seed spec)
		-o, -outDir			The job output directory. Output defined in
										seedfile: Job.Interface.Outputs.Files and
										Job.Interface.Outputs.Json will be stored relative to
										this directory.
		-s, -schema     The Seed Metadata Schema file; Overrides built in schema to validate
									side-car metadata files against

		-rm				Automatically remove the container when it exits (same as
										docker run --rm)
	seed search [OPTIONS]
		Options:
			-r, -registry	The registry to search

			-o, -org		Limit results to this organization/user (for docker hub, this arg is required as images cannot be listed except by org

			-u, -user		Optional username to use for authentication

			-p, -password	Optional password to use for authentication

	seed validate [OPTIONS]
		Options:
			-d, -directory	The directory containing the seed spec
											(default is current directory)
			-s, -schema			Seed Schema file; Overrides built in schema to validate
											spec against.

	seed version
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ngageoint/seed-cli/commands"
	"github.com/ngageoint/seed-cli/constants"
	"github.com/ngageoint/seed-cli/objects"
	"github.com/ngageoint/seed-cli/util"
)

var buildCmd *flag.FlagSet
var initCmd *flag.FlagSet
var listCmd *flag.FlagSet
var publishCmd *flag.FlagSet
var pullCmd *flag.FlagSet
var runCmd *flag.FlagSet
var searchCmd *flag.FlagSet
var validateCmd *flag.FlagSet
var versionCmd *flag.FlagSet
var version string

func main() {
	// Handles any panics/actual exits. Ensures deferred functions are called
	// before program exit.
	defer util.HandleExit()

	//set config dir so we don't stomp on other users' logins with sudo
	os.Setenv("DOCKER_CONFIG", constants.DockerConfigDir)

	// Parse input flags
	DefineFlags()

	// seed init: Create example seed.manifest.json. Does not require docker
	if initCmd.Parsed() {
		dir := initCmd.Lookup(constants.JobDirectoryFlag).Value.String()
		err := commands.SeedInit(dir)
		if err != nil {
			panic(util.Exit{1})
		}
		panic(util.Exit{0})
	}

	// seed validate: Validate seed.manifest.json. Does not require docker
	if validateCmd.Parsed() {
		schemaFile := validateCmd.Lookup(constants.SchemaFlag).Value.String()
		dir := validateCmd.Lookup(constants.JobDirectoryFlag).Value.String()
		err := commands.Validate(schemaFile, dir)
		if err != nil {
			panic(util.Exit{1})
		}
		panic(util.Exit{0})
	}

	// seed search: Searches registry for seed images. Does not require docker
	if searchCmd.Parsed() {
		url := searchCmd.Lookup(constants.RegistryFlag).Value.String()
		org := searchCmd.Lookup(constants.OrgFlag).Value.String()
		filter := searchCmd.Lookup(constants.FilterFlag).Value.String()
		username := searchCmd.Lookup(constants.UserFlag).Value.String()
		password := searchCmd.Lookup(constants.PassFlag).Value.String()
		err := commands.DockerSearch(url, org, filter, username, password)
		if err != nil {
			panic(util.Exit{1})
		}
		panic(util.Exit{0})
	}

	// Checks if Docker requires sudo access. Prints error message if so.
	util.CheckSudo()

	// seed list: Lists all seed compliant images on (default) local machine
	if listCmd.Parsed() {
		_, err := commands.DockerList()
		if err != nil {
			panic(util.Exit{1})
		}
		panic(util.Exit{0})
	}

	// seed build: Build Docker image
	if buildCmd.Parsed() {
		jobDirectory := buildCmd.Lookup(constants.JobDirectoryFlag).Value.String()
		user := searchCmd.Lookup(constants.UserFlag).Value.String()
		pass := searchCmd.Lookup(constants.PassFlag).Value.String()
		err := commands.DockerBuild(jobDirectory, user, pass)
		if err != nil {
			panic(util.Exit{1})
		}
		panic(util.Exit{0})
	}

	// seed run: Runs docker image provided or found in seed manifest
	if runCmd.Parsed() {
		imageName := runCmd.Lookup(constants.ImgNameFlag).Value.String()
		inputs := strings.Split(runCmd.Lookup(constants.InputsFlag).Value.String(), ",")
		settings := strings.Split(runCmd.Lookup(constants.SettingFlag).Value.String(), ",")
		mounts := strings.Split(runCmd.Lookup(constants.MountFlag).Value.String(), ",")
		outputDir := runCmd.Lookup(constants.JobOutputDirFlag).Value.String()
		rmFlag := runCmd.Lookup(constants.RmFlag).Value.String() == constants.TrueString
		metadataSchema := runCmd.Lookup(constants.SchemaFlag).Value.String()
		err := commands.DockerRun(imageName, outputDir, metadataSchema, inputs, settings, mounts, rmFlag)
		if err != nil {
			panic(util.Exit{1})
		}
		panic(util.Exit{0})
	}

	// seed publish: Publishes a seed compliant image
	if publishCmd.Parsed() {
		registry := publishCmd.Lookup(constants.RegistryFlag).Value.String()
		org := publishCmd.Lookup(constants.OrgFlag).Value.String()
		user := searchCmd.Lookup(constants.UserFlag).Value.String()
		pass := searchCmd.Lookup(constants.PassFlag).Value.String()
		origImg := publishCmd.Arg(0)
		jobDirectory := publishCmd.Lookup(constants.JobDirectoryFlag).Value.String()
		deconflict := publishCmd.Lookup(constants.ForcePublishFlag).Value.String() == "false"

		increasePkgMinor := publishCmd.Lookup(constants.PkgVersionMinor).Value.String() ==
			constants.TrueString
		increasePkgMajor := publishCmd.Lookup(constants.PkgVersionMajor).Value.String() ==
			constants.TrueString
		increaseAlgMinor := publishCmd.Lookup(constants.AlgVersionMinor).Value.String() ==
			constants.TrueString
		increaseAlgMajor := publishCmd.Lookup(constants.AlgVersionMajor).Value.String() ==
			constants.TrueString

		err := commands.DockerPublish(origImg, registry, org, user, pass, jobDirectory, deconflict,
			increasePkgMinor, increasePkgMajor, increaseAlgMinor, increaseAlgMajor)
		if err != nil {
			panic(util.Exit{1})
		}
		panic(util.Exit{0})
	}

	// seed pull: Pulls a remote image and tags it as a local image
	if pullCmd.Parsed() {
		imageName := runCmd.Lookup(constants.ImgNameFlag).Value.String()
		registry := pullCmd.Lookup(constants.RegistryFlag).Value.String()
		org := pullCmd.Lookup(constants.OrgFlag).Value.String()
		user := pullCmd.Lookup(constants.UserFlag).Value.String()
		pass := pullCmd.Lookup(constants.PassFlag).Value.String()

		err := commands.DockerPull(imageName, registry, org, user, pass)
		if err != nil {
			panic(util.Exit{1})
		}
		panic(util.Exit{0})
	}
}

//DefineBuildFlags defines the flags for the seed build command
func DefineBuildFlags() {
	// build command flags
	buildCmd = flag.NewFlagSet(constants.BuildCommand, flag.ContinueOnError)
	var directory string
	buildCmd.StringVar(&directory, constants.JobDirectoryFlag, ".",
		"Directory of seed spec and Dockerfile (default is current directory).")
	buildCmd.StringVar(&directory, constants.ShortJobDirectoryFlag, ".",
		"Directory of seed spec and Dockerfile (default is current directory).")

	var user string
	buildCmd.StringVar(&user, constants.UserFlag, "",
		"Optional username to use if dockerfile pulls images from private repository (default is anonymous).")
	buildCmd.StringVar(&user, constants.ShortUserFlag, "",
		"Optional username to use if dockerfile pulls images from private repository (default is anonymous).")

	var password string
	buildCmd.StringVar(&password, constants.PassFlag, "",
		"Optional password if dockerfile pulls images from private repository (default is empty).")
	buildCmd.StringVar(&password, constants.ShortPassFlag, "",
		"Optional password if dockerfile pulls images from private repository (default is empty).")

	// Print usage function
	buildCmd.Usage = func() {
		commands.PrintBuildUsage()
	}
}

//DefineInitFlags defines the flags for the seed init command
func DefineInitFlags() {
	// build command flags
	initCmd = flag.NewFlagSet(constants.InitCommand, flag.ContinueOnError)
	var directory string
	initCmd.StringVar(&directory, constants.JobDirectoryFlag, ".",
		"Directory to place example seed.manifest.json (default is current directory).")
	initCmd.StringVar(&directory, constants.ShortJobDirectoryFlag, ".",
		"Directory to place example seed.manifest.json (default is current directory).")

	// Print usage function
	initCmd.Usage = func() {
		commands.PrintInitUsage()
	}
}

//DefineRunFlags defines the flags for the seed run command
func DefineRunFlags() {
	runCmd = flag.NewFlagSet(constants.RunCommand, flag.ContinueOnError)

	var imgNameFlag string
	runCmd.StringVar(&imgNameFlag, constants.ImgNameFlag, "",
		"Name of Docker image to run")
	runCmd.StringVar(&imgNameFlag, constants.ShortImgNameFlag, "",
		"Name of Docker image to run")

	var inputs objects.ArrayFlags
	runCmd.Var(&inputs, constants.InputsFlag,
		"Defines the full path to any input data arguments")
	runCmd.Var(&inputs, constants.ShortInputsFlag,
		"Defines the full path to input data arguments")

	var settings objects.ArrayFlags
	runCmd.Var(&settings, constants.SettingFlag,
		"Defines the value to be applied to setting")
	runCmd.Var(&settings, constants.ShortSettingFlag,
		"Defines the value to be applied to setting")

	var mounts objects.ArrayFlags
	runCmd.Var(&mounts, constants.MountFlag,
		"Defines the full path to be mapped via mount")
	runCmd.Var(&mounts, constants.ShortMountFlag,
		"Defines the full path to be mapped via mount")

	var outdir string
	runCmd.StringVar(&outdir, constants.JobOutputDirFlag, "",
		"Full path to the algorithm output directory")
	runCmd.StringVar(&outdir, constants.ShortJobOutputDirFlag, "",
		"Full path to the algorithm output directory")

	var rmVar bool
	runCmd.BoolVar(&rmVar, constants.RmFlag, false,
		"Specifying the -rm flag automatically removes the image after executing docker run")

	var metadataSchema string
	runCmd.StringVar(&metadataSchema, constants.SchemaFlag, "",
		"Metadata schema file to override built in schema in validating side-car metadata files")
	runCmd.StringVar(&metadataSchema, constants.ShortSchemaFlag, "",
		"Metadata schema file to override built in schema in validating side-car metadata files")

	// Run usage function
	runCmd.Usage = func() {
		commands.PrintRunUsage()
	}
}

//DefineListFlags defines the flags for the seed list command
func DefineListFlags() {
	listCmd = flag.NewFlagSet("list", flag.ExitOnError)
	listCmd.Usage = func() {
		commands.PrintListUsage()
	}
}

//DefineSearchFlags defines the flags for the seed search command
func DefineSearchFlags() {
	// Search command
	searchCmd = flag.NewFlagSet(constants.SearchCommand, flag.ExitOnError)
	var registry string
	searchCmd.StringVar(&registry, constants.RegistryFlag, "", "Specifies registry to search (default is index.docker.io).")
	searchCmd.StringVar(&registry, constants.ShortRegistryFlag, "", "Specifies registry to search (default is index.docker.io).")

	var org string
	searchCmd.StringVar(&org, constants.OrgFlag, "", "Specifies organization to filter (default is no filter, search all orgs).")
	searchCmd.StringVar(&org, constants.ShortOrgFlag, "", "Specifies organization to filter (default is no filter, search all orgs).")

	var filter string
	searchCmd.StringVar(&filter, constants.FilterFlag, "", "Specifies filter to apply (default is no filter).")
	searchCmd.StringVar(&filter, constants.ShortFilterFlag, "", "Specifies filter to apply (default is no filter).")

	var user string
	searchCmd.StringVar(&user, constants.UserFlag, "", "Specifies username to use for authorization (default is anonymous).")
	searchCmd.StringVar(&user, constants.ShortUserFlag, "", "Specifies username to use for authorization (default is anonymous).")

	var password string
	searchCmd.StringVar(&password, constants.PassFlag, "", "Specifies password to use for authorization (default is empty).")
	searchCmd.StringVar(&password, constants.ShortPassFlag, "", "Specifies password to use for authorization (default is empty).")

	searchCmd.Usage = func() {
		commands.PrintSearchUsage()
	}
}

//DefinePublishFlags defines the flags for the seed publish command
func DefinePublishFlags() {
	publishCmd = flag.NewFlagSet(constants.PublishCommand, flag.ExitOnError)
	var registry string
	publishCmd.StringVar(&registry, constants.RegistryFlag, "", "Specifies registry to publish image to.")
	publishCmd.StringVar(&registry, constants.ShortRegistryFlag, "", "Specifies registry to publish image to.")

	var org string
	publishCmd.StringVar(&org, constants.OrgFlag, "", "Specifies organization to publish image to.")
	publishCmd.StringVar(&org, constants.ShortOrgFlag, "", "Specifies organization to publish image to.")

	var d string
	publishCmd.StringVar(&d, constants.JobDirectoryFlag, ".",
		"Directory of seed spec and Dockerfile (default is current directory).")
	publishCmd.StringVar(&d, constants.ShortJobDirectoryFlag, ".",
		"Directory of seed spec and Dockerfile (default is current directory).")

	var b bool
	publishCmd.BoolVar(&b, constants.ForcePublishFlag, false,
		"Force publish, do not deconflict")
	var pMin bool
	publishCmd.BoolVar(&pMin, constants.PkgVersionMinor, false,
		"Minor version bump of 'packageVersion' in manifest on disk, will auto rebuild and push")
	var pMaj bool
	publishCmd.BoolVar(&pMaj, constants.PkgVersionMajor, false,
		"Major version bump of 'packageVersion' in manifest on disk, will auto rebuild and push")
	var aMin bool
	publishCmd.BoolVar(&aMin, constants.AlgVersionMinor, false,
		"Minor version bump of 'algorithmVersion' in manifest on disk, will auto rebuild and push")
	var aMaj bool
	publishCmd.BoolVar(&aMaj, constants.AlgVersionMajor, false,
		"Major version bump of 'algorithmVersion' in manifest on disk, will auto rebuild and push")

	publishCmd.Usage = func() {
		commands.PrintPublishUsage()
	}
}

//DefinePullFlags defines the flags for the seed pull command
func DefinePullFlags() {
	// Search command
	pullCmd = flag.NewFlagSet(constants.PullCommand, flag.ExitOnError)

	var imgNameFlag string
	pullCmd.StringVar(&imgNameFlag, constants.ImgNameFlag, "",
		"Name of Docker image to pull")
	pullCmd.StringVar(&imgNameFlag, constants.ShortImgNameFlag, "",
		"Name of Docker image to pull")

	var registry string
	pullCmd.StringVar(&registry, constants.RegistryFlag, "", "Specifies registry to pull image from (default is index.docker.io).")
	pullCmd.StringVar(&registry, constants.ShortRegistryFlag, "", "Specifies registry to pull image from (default is index.docker.io).")

	var org string
	pullCmd.StringVar(&org, constants.OrgFlag, "", "Specifies organization to pull image from (default is geoint).")
	pullCmd.StringVar(&org, constants.ShortOrgFlag, "", "Specifies organization to pull image from (default is geoint).")

	var user string
	pullCmd.StringVar(&user, constants.UserFlag, "", "Specifies username to use for authorization (default is anonymous).")
	pullCmd.StringVar(&user, constants.ShortUserFlag, "", "Specifies username to use for authorization (default is anonymous).")

	var password string
	pullCmd.StringVar(&password, constants.PassFlag, "", "Specifies password to use for authorization (default is empty).")
	pullCmd.StringVar(&password, constants.ShortPassFlag, "", "Specifies password to use for authorization (default is empty).")

	pullCmd.Usage = func() {
		commands.PrintPullUsage()
	}
}

//DefineValidateFlags defines the flags for the validate command
func DefineValidateFlags() {
	var directory string
	validateCmd = flag.NewFlagSet(constants.ValidateCommand, flag.ExitOnError)
	validateCmd.StringVar(&directory, constants.JobDirectoryFlag, ".",
		"Location of the seed.manifest.json spec to validate")
	validateCmd.StringVar(&directory, constants.ShortJobDirectoryFlag, ".",
		"Location of the seed.manifest.json spec to validate")
	var schema string
	validateCmd.StringVar(&schema, constants.SchemaFlag, "",
		"JSON schema file to validate seed against.")
	validateCmd.StringVar(&schema, constants.ShortSchemaFlag, "",
		"JSON schema file to validate seed against.")

	validateCmd.Usage = func() {
		commands.PrintValidateUsage()
	}
}

//DefineFlags defines the flags available for the seed runner.
func DefineFlags() {
	// Seed subcommand flags
	DefineBuildFlags()
	DefineInitFlags()
	DefineRunFlags()
	DefineListFlags()
	DefineSearchFlags()
	DefinePublishFlags()
	DefineValidateFlags()
	versionCmd = flag.NewFlagSet(constants.VersionCommand, flag.ExitOnError)
	versionCmd.Usage = func() {
		PrintVersionUsage()
	}

	// Print usage if no command given
	if len(os.Args) == 1 {
		PrintUsage()
	}

	var cmd *flag.FlagSet = nil
	minArgs := 2

	// Parse commands
	switch os.Args[1] {
	case constants.BuildCommand:
		cmd = buildCmd

		// Check for seed manifest in current directory. If found, add current directory arg
		if len(os.Args) == 2 {
			if _, err := util.SeedFileName("."); err == nil {
				os.Args = append(os.Args, ".")
			}
		}
		minArgs = 3

	case constants.InitCommand:
		cmd = initCmd
		minArgs = 2

	case constants.RunCommand:
		cmd = runCmd
		minArgs = 3

	case constants.SearchCommand:
		cmd = searchCmd
		minArgs = 2

	case constants.ListCommand:
		cmd = listCmd
		minArgs = 2

	case constants.PublishCommand:
		cmd = publishCmd
		minArgs = 3

	case constants.ValidateCommand:
		cmd = validateCmd
		minArgs = 3

	case constants.VersionCommand:
		versionCmd.Parse(os.Args[2:])
		PrintVersion()

	default:
		fmt.Fprintf(os.Stderr, "%q is not a valid command.\n", os.Args[1])
		PrintUsage()
		panic(util.Exit{0})
	}

	if cmd != nil {
		cmd.Parse(os.Args[2:])
		if len(os.Args) < minArgs {
			cmd.Usage()
		}
	}
}

//PrintUsage prints the seed usage arguments
func PrintUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed COMMAND\n\n")
	fmt.Fprintf(os.Stderr, "A test runner for seed spec compliant algorithms\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  build \tBuilds Seed compliant Docker image\n")
	fmt.Fprintf(os.Stderr, "  init \tInitialize new project with example seed.manifest.json file\n")
	fmt.Fprintf(os.Stderr, "  list  \tAllows for listing of all Seed compliant images residing on the local system\n")
	fmt.Fprintf(os.Stderr, "  publish\tAllows for publish of Seed compliant images to remote Docker registry\n")
	fmt.Fprintf(os.Stderr, "  run   \tExecutes Seed compliant Docker docker image\n")
	fmt.Fprintf(os.Stderr, "  search\tAllows for discovery of Seed compliant images hosted within a Docker registry (default is docker.io)\n")
	fmt.Fprintf(os.Stderr, "  validate\tValidates a Seed spec\n")
	fmt.Fprintf(os.Stderr, "  version\tPrints the version of Seed spec\n")
	fmt.Fprintf(os.Stderr, "\nRun 'seed COMMAND --help' for more information on a command.\n")
	panic(util.Exit{0})
}

//PrintVersionUsage prints the seed version usage, then exits the program
func PrintVersionUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed version \n")
	fmt.Fprintf(os.Stderr, "\nOutputs the version of the Seed CLI and specification.\n")
	panic(util.Exit{0})
}

//PrintVersion prints the seed CLI version
func PrintVersion() {
	fmt.Fprintf(os.Stderr, "Seed v%s\n", version)
	panic(util.Exit{0})
}
