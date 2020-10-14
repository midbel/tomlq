# query

query has been created in order to make easier to retrieve values from a [toml document](https://toml.io) like [jq](https://stedolan.github.io/jq/) does with JSON document. In some extend, query can also be used to search for values in JSON documents.

### Syntax

query has a little syntax based on the toml specification but also extended in order to test for the presence of an option and/or its expected values.

It is also inspired by the syntax of the CSS selectors in order to select elements of array or to check the type of a value (array, table,...).

There are other influences such as the notation of regular expression in Javascript, the way to specified the kind of variable in Perl.

The general form of a query is as follows:

```
[level][type]element[:selector][[predicate]][subquery],[query]...
```

##### level

The level part of the query specifies where the query should try to find the requested element.

There are two levels defined:

1. Level **One**: the query will only try to find the given element in the current level - most of the time the current table. If it can not find it, it will not try to go deeper in the document and returns a null result.
2. Level **Any**: the query will try to find the given element in the current level first. If it can not find it, it will start to check the sub tables until it can find the specified element. That also means that as soon as the query find a match, even if other match are possible deeper in the document, they will be ignored.

The level **One** is written with the single dot operator: ```.```

The level **Any** is written with the double dot operator: ```..```

Note: the level operator is optional in the query. If not specified, the query will considered that level **Any** has been specified.

##### type

The type part of the query allows to specify the type of elements that the query should search for. By default, the query will search for any types of element. When a list of elements (see below) is given, all elements in this list will have to be of the same type.

There are three types that are defined and that match (more or less) the three types of elements found in toml document:

* Value: the query will only look for simple values. Simple values are string, integer, float, boolean and date and time values.
* Regular: the query will only look for tables in the document. It does not matter if table are defined inline or as member of an array or regular table.
* Array: the query will only look for array whatever the kind of elements they contain.

The type **Value** is written with the percent operator: ```%```

The type **Regular** is written with the dollar operator: ```$```

The type **Array** is written with the arobase operator: ```@```

As described above, if none of these operators appears in the query, the query will look for any type of elements.

##### element

The element part of the query specify the **key** of an element to look for in the document. The key can be the name of a table or the name of an option.

The rules to specify the name of key in the document are more or less identical at the one described in the toml specification:

* bare key(s) allows only alphanumeric, dash and underscore characters. Notes that the first character should be a letter. A bare key can also be an integer but should only be composed of digits.
* quoted key(s) follow the same rules as basic strings (surrounded by quotation mark ```"```) or literal strings (surrounded by single quote ```'```). Into a quoted key, all the special characters are escaped and loses their special meaning

query introduce also the possibility to find an element using a simple pattern similar to the one used to glob files in Linux shell. Into a pattern, all the special characters are escaped and loses their special meaning

To specify a pattern in the query (instead of a "regular" key), the pattern should be surrounded by slash ```/```.

The syntax of a pattern is made of the following elements:

* ```*```: match zeros or any sequence of characters in the input
* ```?```: match any single character in the input
* ```[]```: list and/or range of character to match a character in the input
* ```!, ^```: negate the match of the list/range characters
* ```\```: used to escape the special meaning of the ```*```, ```?```, ```[```, ```\``` characters

Moroever, query does not limit to select one element per query. Indeed, it is also possible to specify a list of elements that the query should match.

Some examples:

```
# a bare key
key

# a integer key
1234

# a quoted key
"key"
'key'

# a pattern
/[A-Z]??b[a-z][a-z]@*.[A-Za-z][A-Za-z][A-Za-z]/

# a list of elements (mix between bare key, integer and pattern)
(key, "key", 1234, /???*/)
```

##### :selector

##### [predicate]

##### subquery

##### Examples

with this sample document:

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

1. getting the repository of the project. We specify that the repository value should be an option with a simple value (not an array nor a table):
```
.%repository
```

2. getting all the changelog that happens after septembre 2020:
```
.changelog[date >= 2020-10-01]
```

3. getting the date of all the changelog that happens after septembre 2020:
```
.changelog[date >= 2020-10-01].date
```

4. getting the first dependency of the project. We specify that the dependency table should be an array:
```
..@dependency:first
```

5. getting the first dependency of the project with the "at" selector and that table has the "version" option present and the optional option is equal to false:
```
..@dependency:at(0)[version && optional == false]
```

6. getting the repository and all the dependencies:
```
.(repository, dependency)
```

7. last example can also be expressed in the following way. The previous version does not allow to set specific type of the targeted elements and/or selectors and/or predicates:
```
.repository,.@dependency:range(,5)[optional == true]
```

### Possible improvements - things to do:

* support for functions
* comparing value of an option with value of another option somewhere in the same document
* specify the root element (table, array) from where the query will be executed
* greedy query
