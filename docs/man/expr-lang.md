# Expr - a powerful expression language

*Heimdall* supports the same expression language, which is also used by
[Argo Rollouts and Argo Workflows](https://argoproj.github.io/) for Kubernetes.
The official [Language Definition](https://expr.medv.io/docs/Language-Definition) provides a good overview.

In addition to that, *Heimdall* registers more than [120 functions](https://go-task.github.io/slim-sprig/) e.g.,

* [String Functions](https://go-task.github.io/slim-sprig/strings.html): `trim`, `plural`, etc.
  * [String List Functions](https://go-task.github.io/slim-sprig/string_slice.html): `splitList`, `sortAlpha`, etc.
* [Integer Math Functions](https://go-task.github.io/slim-sprig/math.html): `add`, `max`, `mul`, etc.
  * [Integer Array Functions](https://go-task.github.io/slim-sprig/integer_slice.html): `until`, `untilStep`, `seq`
* [Date Functions](https://go-task.github.io/slim-sprig/date.html): `now`, `date`, etc.
* [Defaults Functions](https://go-task.github.io/slim-sprig/defaults.html):
  `default`, `empty`, `coalesce`, `fromJson`, `toJson`, `toPrettyJson`, `toRawJson`, `ternary`
* [Encoding Functions](https://go-task.github.io/slim-sprig/encoding.html): `b64enc`, `b64dec`, etc.
* [Lists and List Functions](https://go-task.github.io/slim-sprig/lists.html): `list`, `first`, `uniq`, etc.
* [Dictionaries and Dict Functions](https://go-task.github.io/slim-sprig/dicts.html):
  `get`, `set`, `dict`, `hasKey`, `pluck`, `dig`, etc.
* [Type Conversion Functions](https://go-task.github.io/slim-sprig/conversion.html): `atoi`, `int64`, `toString`, etc.
* [Path and Filepath Functions](https://go-task.github.io/slim-sprig/paths.html):
  `base`, `dir`, `ext`, `clean`, `isAbs`, `osBase`, `osDir`, `osExt`, `osClean`, `osIsAbs`
* [Flow Control Functions](https://go-task.github.io/slim-sprig/flow_control.html): `fail`
* Advanced Functions
  * [OS Functions](https://go-task.github.io/slim-sprig/os.html): `env`, `expandenv`
  * [Reflection](https://go-task.github.io/slim-sprig/reflection.html): `typeOf`, `kindIs`, `typeIsLike`, etc.
  * [Cryptographic and Security Functions](https://go-task.github.io/slim-sprig/crypto.html): `sha256sum`, etc.
