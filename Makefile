.PHONY: package
package: lets-goooo.zip

lets-goooo.zip: doc.pdf
	rm -f lets-goooo.zip
	zip lets-goooo.zip src/assets src/certification src/cmd src/pkg src/template src/go.mod src/go.sum src/locations.xml src/logoooo.png doc.pdf

.PHONY: doc.pdf
doc.pdf:
	rm -rf tmp
	gh run download `gh workflow view "Deploy Pandoc" | sed -n "s/.*push\s*\S*\s*\([0-9][0-9]*\).*/\1/p" | head -n 1` -n documentation --dir tmp
	mv tmp/doc.pdf doc.pdf
	rm -rf tmp
