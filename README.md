### Usage

You can find executables for Windows, Linux and macOS (Darwin) in the `./bin` directory. The executable will be referred to as `./spt` throughout this document. 

Run `./spt --help` for usage:

```
Usage of ./spt:
    -consonant string
      	path to file with consonant encodings
    -cost_groups_plot string
      	path to file with cost groups plot
    -count_groups_plot string
      	path to file with count groups plot
    -lang_1 string
      	first language to compare (optional)
    -lang_2 string
      	second language to compare (optional)
    -num_trials int
      	number of trials (default 1000000)
    -output string
      	path to output file (stdout if not specified)
    -set_a string
      	path to file containing wordlists for A (triggers AB mode)
    -set_b string
      	path to file containing wordlists for B (triggers AB mode)
    -sounds string
      	path to file containing sound classes (default "./data/sounds.xlsx")
    -verbose
      	verbose output
    -weights string
      	path to file containing class weights
    -wordlists string
      	path to file containing wordlists (default "./data/wordlists.xlsx")
```

##### Running test on two wordlists

To run test on a pair of wordlists contained in a single file, run:

```
$ ./spt --num_trials=1000000 --sounds=./data/sounds.xlsx --wordlists=./data/wordlists.xlsx --weights=./data/weights.xlsx
n = 50 (number of compared pairs)

Positive pair 0: 19 drink: eːgʰʷ - ɨɣi
Positive pair 1: 39 hear: kʸlew - kuwli
Positive pair 2: 42 I : me - mi
Positive pair 3: 57 name: nom - nimi
Positive pair 4: 87 thou : ti - ti
Positive pair 5: 94 water: wed - weti
Positive pair 6: 98 who: kʷi - ku
N = 7 (number of positive pairs in the original list)
S = 436.000000 (cost of positive pairs in the original list)

k = 0:	89679 trial(s)
k = 1:	231915 trial(s)
k = 2:	281519 trial(s)
k = 3:	214596 trial(s)
k = 4:	115495 trial(s)
k = 5:	46891 trial(s)
k = 6:	15108 trial(s)
k = 7:	3863 trial(s)
k = 8:	768 trial(s)
k = 9:	148 trial(s)
k = 10:	17 trial(s)
k = 11:	1 trial(s)
P (counts) = 4797 / 1000000 = 0.004797

s = 0.000: 89679 trial(s)
s = 37.000: 24056 trial(s)
s = 39.000: 3864 trial(s)
s = 41.000: 20597 trial(s)
<...>
s = 574.000: 1 trial(s)
s = 580.000: 1 trial(s)
s = 600.000: 1 trial(s)
s = 601.000: 1 trial(s)
P (costs) = 681 / 1000000 = 0.000681
```

* `--num_trials` specifies how many times we shuffle the wordlists and count scores; default value is `1000000`.
* `--sounds` is the path to file with sound tables; sample file can be found at `./data/sounds.xlsx` (also the default value).
* `--wordlists` is the path to file with wordlists; sample file can be found at `./data/wordlists.xlsx` (also the default value).
* `--weights` is the path containing mapping from Swadesh ID to its weight (missing IDs get weight value of 1.0); sample file can be found at `./data/weights.xlsx`.

##### Running test on two sets of wordlists (AB mode)

```
$ ./spt --num_trials=1000000 --sounds=./data/sounds.xlsx --set_a=./data/wordlists.xlsx --set_b=./data/wordlists.xlsx
```

* `--num_trials` specifies how many times we shuffle the wordlists and count scores; default value is `1000000`.
* `--sounds` is the path to file with sound tables; sample file can be found at `./data/sounds.xlsx` (also the default value).
* `--set_a` is the path to file with wordlists set A (no default value).
* `--set_b` is the path to file with wordlists set B (no default value).
* `--weights` is the path containing mapping from Swadesh ID to its weight (missing IDs get weight value of 1.0); sample file can be found at `./data/weights.xlsx` (also the default value).

If you run the command as specified above (using the same file for sets A and B), the program will execute correctly (but probability will always be zero, which is expected).

##### Building plots

Pass the `--count_groups_plot` option to build a plot representing how many trials gave a certain amount of matches:

```
$ ./spt --count_groups_plot=./count_groups.svg
``` 

Sample count groups plot:

![Sample plot](https://github.com/oopcode/starling-permutation-test/blob/master/count_groups.svg)

Pass the `--cost_groups_plot` option to build a plot representing how many trials gave a certain cost of matches:

```
$ ./spt --cost_groups_plot=./cost_groups.svg
``` 

Sample cost groups plot:

![Sample plot](https://github.com/oopcode/starling-permutation-test/blob/master/cost_groups.svg)

### Building from source

You need to have [Go](https://golang.org/doc/install) installed. Project must be put to `$GOPATH/src/github.com/`. Run:

```
make deps
make build
```