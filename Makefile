SRCS := $(wildcard *.md)
OBJS := $(patsubst %.md,%.html,$(SRCS))

all: $(OBJS)

%.html: %.md
	# Use -i to create incremental lists
	# http://pages.stat.wisc.edu/~yandell/statgen/ucla/Help/Producing%20slide%20shows%20with%20Pandoc.html
	pandoc -s \
		--webtex \
		-t revealjs \
		--slide-level=2 \
		-V revealjs-url='https://cdnjs.cloudflare.com/ajax/libs/reveal.js/3.9.2' \
		-V theme=league \
		$< -o $@

