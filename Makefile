.PHONY: package
package: lets-goooo.zip

lets-goooo.zip: doc.pdf
	rm -f lets-goooo.zip
	zip -r lets-goooo.zip assets certification cmd internal template go.mod go.sum locations.xml logoooo.png doc.pdf

.PHONY: doc.pdf
doc.pdf:
	rm -rf tmp
	gh run download `gh workflow view "Deploy Pandoc" | sed -n "s/.*push\s*\S*\s*\([0-9][0-9]*\).*/\1/p" | head -n 1` -n documentation --dir tmp
	mv tmp/doc.pdf doc.pdf
	rm -rf tmp
