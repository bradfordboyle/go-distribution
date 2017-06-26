#!/bin/sh

# make sure env is setup proper
if [ ! -x "./distribution" ] ; then
	go build ..
fi

distribution="./distribution"

getopts "v" verbose

# the tests
echo ""
printf "Running test: 1. "
cat stdin.01.txt | $distribution --rcfile=../distributionrc --graph --height=35 --width=120 --char=dt --color --verbose > stdout.01.actual.txt 2> stderr.01.actual.txt

echo "done."

# be sure output is proper
err=0
printf "Comparing results: "
for i in 01 ; do
	printf "$i. \n"
	diff -w stdout.$i.expected.txt stdout.$i.actual.txt
	if [ $? -ne 0 ]; then
		err=1
	fi

	# when in verbose mode, ignore any "runtime lines, since those may differ by
	# milliseconds from machine to machine. Also ignore any lines with "^M" markers,
	# which are line-erase signals used for updating the screen interactively, and
	# thus don't need to be stored or compared.
	if [ "$verbose" = "v" ]; then
		diff -w -I "runtime:" -I "" stderr.$i.expected.txt stderr.$i.actual.txt
	fi
done

echo "done."

# clean up
rm stdout.*.actual.txt stderr.*.actual.txt $distribution

exit $err
