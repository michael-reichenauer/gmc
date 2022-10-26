#tool nuget:?package=NUnit.ConsoleRunner&version=3.4.0
#tool nuget:?package=Tools.InnoSetup&version=5.5.9
#tool nuget:?package=Microsoft.VSSDK.Vsixsigntool&version=15.7.27703
#addin nuget:?package=Cake.VersionReader&version=5.0.0
#addin nuget:?package=Cake.VsixSignTool&version=1.2.0

//////////////////////////////////////////////////////////////////////
// ARGUMENTS
//////////////////////////////////////////////////////////////////////

var target = Argument("target", "Default");
var configuration = Argument("configuration", "Release");


//////////////////////////////////////////////////////////////////////
// PREPARATION
//////////////////////////////////////////////////////////////////////

// Define paths.
var name = "gmc";

var outputPath = $"gmc.exe";
var setupPath = $"{name}Setup.exe";
var uninstallerPath = $"Setup/Sign/Uninstaller.exe";
var signedUninstallerPath = $"Setup/Sign/uninst-5.5.9 (u)-44666f8110.e32";



//////////////////////////////////////////////////////////////////////
// TASKS
//////////////////////////////////////////////////////////////////////

Task("Clean")
    .Does(() =>
{
    CleanDirectories($"./**/obj/{configuration}");
    CleanDirectories($"./**/bin/{configuration}");

    if (FileExists(setupPath))
    { 
        DeleteFile(setupPath);
    }
});


Task("Build-Setup-File")
    .Does(() =>
{
    //var version = GetFullVersionNumber(outputPath);
    var version = "0.58";
	string isSigning = "False";

	InnoSetup("./Setup/gmc.iss", new InnoSetupSettings {
		QuietMode = InnoSetupQuietMode.QuietWithProgress,
		Defines = new Dictionary<string, string> { 
			{"AppVersion", ""},
			{"ProductVersion", version},
			{"IsSigning", isSigning},
		}
    });
});


Task("Build-Setup")
    .IsDependentOn("Clean")
    .IsDependentOn("Build-Setup-File")
    .Does(() =>
{  
})
.OnError(exception =>
{
	RunTarget("Clean");
	throw exception;
});;



Task("Default");



//////////////////////////////////////////////////////////////////////
// EXECUTION
//////////////////////////////////////////////////////////////////////

RunTarget(target);
