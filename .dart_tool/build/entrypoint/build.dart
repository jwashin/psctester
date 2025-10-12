// @dart=3.6
// ignore_for_file: directives_ordering
// build_runner >=2.4.16
// ignore_for_file: no_leading_underscores_for_library_prefixes
import 'package:build_runner/src/build_plan/builder_factories.dart' as _i1;
import 'package:build_modules/builders.dart' as _i2;
import 'package:build_test/builder.dart' as _i3;
import 'package:build_web_compilers/builders.dart' as _i4;
import 'dart:io' as _i5;
import 'package:build_runner/src/bootstrap/processes.dart' as _i6;

final _builderFactories = _i1.BuilderFactories(
  builderFactories: {
    'build_modules:module_library': [_i2.moduleLibraryBuilder],
    'build_test:test_bootstrap': [
      _i3.debugIndexBuilder,
      _i3.debugTestBuilder,
      _i3.testBootstrapBuilder,
    ],
    'build_web_compilers:dart2js_modules': [
      _i4.dart2jsMetaModuleBuilder,
      _i4.dart2jsMetaModuleCleanBuilder,
      _i4.dart2jsModuleBuilder,
    ],
    'build_web_compilers:dart2wasm_modules': [
      _i4.dart2wasmMetaModuleBuilder,
      _i4.dart2wasmMetaModuleCleanBuilder,
      _i4.dart2wasmModuleBuilder,
    ],
    'build_web_compilers:ddc': [
      _i4.ddcKernelBuilder,
      _i4.ddcBuilder,
    ],
    'build_web_compilers:ddc_modules': [
      _i4.ddcMetaModuleBuilder,
      _i4.ddcMetaModuleCleanBuilder,
      _i4.ddcModuleBuilder,
    ],
    'build_web_compilers:entrypoint': [_i4.webEntrypointBuilder],
    'build_web_compilers:sdk_js': [
      _i4.sdkJsCompile,
      _i4.sdkJsCopyRequirejs,
    ],
  },
  postProcessBuilderFactories: {
    'build_modules:module_cleanup': _i2.moduleCleanup,
    'build_web_compilers:dart2js_archive_extractor':
        _i4.dart2jsArchiveExtractor,
    'build_web_compilers:dart_source_cleanup': _i4.dartSourceCleanup,
  },
);
void main(List<String> args) async {
  _i5.exitCode = await _i6.ChildProcess.run(
    args,
    _builderFactories,
  )!;
}
