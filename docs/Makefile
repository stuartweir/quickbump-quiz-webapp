MARKDOWN = markdown

all: data.html http.html

clean:
	rm *.html

%.html: %.md
	echo '<link href="http://kevinburke.bitbucket.org/markdowncss/markdown.css" rel="stylesheet"></link>' > $@
	$(MARKDOWN) $< >> $@
