# query

query has been created in order to make easier to retrieve values from a [toml document](https://toml.io) like [jq](https://stedolan.github.io/jq/) does with JSON document. In some extend, query can also be used to search for values in JSON documents.

### query syntax

query has a little syntax based on the toml specification but also extended in order to test for the presence of an option and/or its expected values. It is also inspired by the syntax of the CSS selectors in order to select elements of array or to check the type of a value (array, table,...).

The general form of a query is:

```
[level][type]label[:selector][[expression]][subquery],[query]...
```

### examples

```toml
package    = "query"
repository = "https://github.com/midbel/query"
language   = "go"

[maintainer]
name = "midbel"
email = "no-reply@midbel.org"

[[changelog]]
date = 2020-10-13 21:35:10+02:00
desc = "edit README with query examples"

[[changelog]]
date = 2020-10-12 09:00:00+02:00
desc = "write tests"

[[changelog]]
date = 2020-10-05 09:00:00+02:00
desc = "write scanner and parser of query language"

[[changelog]]
date = 2020-10-03 16:30:00+02:00
desc = "first draft of query language"

[[changelog]]
date = 2020-09-01 09:00:00+02:00
desc = "bash script using grep and sed to retrieve value from toml document"

[[dependency]]
package = "toml"
repository = "https://github.com/midbel/toml"
version    = "0.1.0"
optional   = false
```

getting the repository of the project. We specify that the repository value should be an option with a simple value (not an array nor a table):
```
.%repository
```

getting all the changelog that happens after septembre 2020:
```
.changelog[date >= 2020-10-01]
```

getting the date of all the changelog that happens after septembre 2020:
```
.changelog[date >= 2020-10-01].date
```

getting the first dependency of the project. We specify that the dependency table should be an array:
```
..@dependency:first
```

getting the first dependency of the project with the "at" selector and that table has the "version" option present and the optional option is equal to false:
```
..@dependency:at(0)[version && optional == false]
```

getting the repository and all the dependencies:
```
.(repository, dependency)
```

last example can also be expressed in the following way. The previous version does not allow to set specific type of the targeted elements and/or selectors and/or predicates:
```
.repository,.@dependency:range(,5)[optional == true]
```

### improvements - things to do:

* support for functions
* comparing value of an option with value of another option somewhere in the same document
