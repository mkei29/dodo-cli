import { spawnSync } from "node:child_process";
import { createRequire } from "node:module";

export function main() {
    callCmd(process.argv.slice(2));
}

export function callCmd(args) {
    const binaryPath = getBinaryPath();
    if (!binaryPath) {
        console.error("Unsupported platform or architecture");
        process.exit(1);
    }
    const result = spawnSync(binaryPath, args, {
        shell: false,
        stdio: "inherit",
    });
    if (result.error) {
        throw result.error;
    }
    process.exitCode = result.status || 0;
}

function getBinaryPath() {
    const DODO_BINARY_PATH = process.env.DODO_BINARY_PATH;
    if (DODO_BINARY_PATH) {
        return DODO_BINARY_PATH;
    }
    const availablePlatforms = ["linux", "darwin"];
    const availableArchs = ["x86-64", "arm64"];
    let { platform, arch } = process;
    arch = rewriteArch(arch);

    if (!availablePlatforms.includes(platform) || !availableArchs.includes(arch)) {
        return;
    }
    const binaryName = `@dodo-doc/cli-${platform}-${arch}/dodo-cli`;
    const require = createRequire(import.meta.url);
    return require.resolve(binaryName);
}

function rewriteArch(arch) {
    if (arch === "x64") {
        return "x86-64";
    }
    return arch;
}
