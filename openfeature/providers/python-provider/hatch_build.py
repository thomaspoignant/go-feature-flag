"""
Hatchling build hook for the GO Feature Flag Python provider.

The provider ships a WASI evaluation binary (``gofeatureflag-evaluation_<version>.wasi``)
for in-process flag evaluation. The binaries live in the ``wasm-releases`` git submodule
(excluded from the distribution) and are git-ignored inside the package, so this hook
copies the pinned-version binary into ``gofeatureflag_python_provider/wasm/`` before the
wheel/sdist is built.

The pinned version is the single source of truth stored in
``gofeatureflag_python_provider/wasm/_wasi_version.txt``.
"""

import shutil
from pathlib import Path

from hatchling.builders.hooks.plugin.interface import BuildHookInterface


class CustomBuildHook(BuildHookInterface):
    """Copy the pinned WASI binary from the wasm-releases submodule into the package."""

    def initialize(self, version: str, build_data: dict) -> None:
        root = Path(self.root)
        wasm_dir = root / "gofeatureflag_python_provider" / "wasm"
        wasm_version = (
            (wasm_dir / "_wasi_version.txt").read_text(encoding="utf-8").strip()
        )
        if not wasm_version:
            raise ValueError(
                f"Empty WASI version in {wasm_dir / '_wasi_version.txt'}; "
                "it must contain the pinned version (e.g. 0.2.3)."
            )
        filename = f"gofeatureflag-evaluation_{wasm_version}.wasi"

        # Remove any stale binaries for other versions so only the pinned one is
        # packaged (the artifacts glob would otherwise ship every *.wasi present).
        for stale in wasm_dir.glob("gofeatureflag-evaluation_*.wasi"):
            if stale.name != filename:
                stale.unlink()

        dest = wasm_dir / filename
        if dest.exists():
            # Already present (e.g. copied by a previous build) - nothing to do.
            return

        source = root / "wasm-releases" / "evaluation" / filename
        if not source.exists():
            raise FileNotFoundError(
                f"WASI binary {filename!r} not found at {source}. "
                "Initialize the submodule first: "
                "`git submodule update --init openfeature/providers/python-provider/wasm-releases`."
            )

        shutil.copyfile(source, dest)
