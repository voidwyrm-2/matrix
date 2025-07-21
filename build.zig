const std = @import("std");
const sep = std.fs.path.sep_str;

const builtin = @import("builtin");

const exe_name = "matrix";

const pkg_folder = "build";

const GoOsTag = enum {
    darwin,
    linux,
    windows,
};

const GoArchTag = enum {
    arm64,
    amd64,
};

const targets = [_]struct { os: GoOsTag, arch: GoArchTag }{
    .{ .os = .darwin, .arch = .arm64 },
    .{ .os = .darwin, .arch = .amd64 },
    .{ .os = .linux, .arch = .arm64 },
    .{ .os = .linux, .arch = .amd64 },
    .{ .os = .windows, .arch = .arm64 },
    .{ .os = .windows, .arch = .amd64 },
};

extern fn putenv(string: [*:0]u8) c_int;

const CEnvError = error{
    PutEnvError,
};

pub fn build(b: *std.Build) !void {
    const package = b.step("pack", "Build and package the Go project");

    const rm_pkg = b.addSystemCommand(&.{ "rm", "-rf", pkg_folder });

    const mkdir_pkg = b.addSystemCommand(&.{ "mkdir", pkg_folder });
    mkdir_pkg.step.dependOn(&rm_pkg.step);

    for (targets) |t| {
        //std.process.getEnvVarOwned(b.allocator, "")

        const real_exe_name = try std.fmt.allocPrint(b.allocator, "{s}{s}", .{ exe_name, if (t.os == .windows) ".exe" else "" });

        {
            const str = try std.fmt.allocPrint(b.allocator, "GOOS={s}{c}", .{ @tagName(t.os), 0 });

            const err = putenv(@ptrCast(str.ptr));

            if (err != 0)
                return CEnvError.PutEnvError;
        }

        {
            const str = try std.fmt.allocPrint(b.allocator, "GOARCH={s}{c}", .{ @tagName(t.arch), 0 });

            const err = putenv(@ptrCast(str.ptr));

            if (err != 0)
                return CEnvError.PutEnvError;
        }

        const build_cmd = &.{
            "go",
            "build",
            "-o",
            real_exe_name,
            ".",
        };

        const build_exe = b.addSystemCommand(build_cmd);

        const zip_cmd = &.{
            "zip",
            "-r",
            try std.fmt.allocPrint(b.allocator, "{s}{s}{s}-{s}.zip", .{ pkg_folder, sep, @tagName(t.os), @tagName(t.arch) }),
            real_exe_name,
        };

        const zip = b.addSystemCommand(zip_cmd);

        zip.step.dependOn(&mkdir_pkg.step);
        zip.step.dependOn(&build_exe.step);

        package.dependOn(&zip.step);
    }
}
