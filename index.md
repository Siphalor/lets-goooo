---
title: Let's Goooo Dokumentation
author: 1103207, 3106335, 4485500
date: \today
lang: de-DE
documentclass: scrreport
papersize: a4
linestretch: 1.5
---

\pagebreak

# Architekturdokumentation

![UML-Klassendiagramm für das gesamte Projekt](img/plantuml/all.png){#fig:uml-all}

In @fig:uml-all ist ein Klassendiagramm für das gesamte Projekt zu sehen.

## Journal

![UML-Klassendiagramm des `journal`-Packages](img/plantuml/pkg_journal.png)

Das Journal-Paket ist in zwei wichtige Klassen geteilt: den `journal.Writer` und den `journal.Reader`.

Ersterer wird vom Server verwendet um die Logins und Logouts zu protokollieren.
Der `Reader` wird hauptsächlich im Analyzer eingesetzt, um die Journal-Dateien einheitlich einzulesen.

### Dateiformat

Die Journal-Dateien werden in einem eigenen Text-Datei-Format gespeichert.

Jede Zeile beginnt dabei mit einem Symbol, dass den Typ dieser Zeile markiert:

- `*`: Ein neuer Nutzer tritt erstmals in der Journaldatei auf. In dieser Zeile werden Name und Adresse durch einen Tab getrennt gespeichert.
- `+`: Ein Nutzer meldet sich an einem Standort an. Gespeichert wird der Nutzer, der Standort und ein Unix-Timestamp des Ereignisses.
- `-`: Ein Nutzer meldet sich an einem Standort ab. Analog zur `+`-Zeile.

Nur in `*`-Zeilen werden die Nutzerdaten direkt gespeichert.
Um Speicherplatz zu sparen wird in folgenden Zeilen der Nutzer nur noch über einen Hashwert seines Namens und seiner Adresse referenziert.+

### Beispiel-Datei
```
*Tester	Teststadt
+HjLV+aPwKzq3szuae53Zv5n4puw=	TST	1634700000
*Klaus	Musterdorf
+O+Dig24BxOFwjJEN1oBbk/VW/tA=	HST	1634710000
-HjLV+aPwKzq3szuae53Zv5n4puw=	TST	1634711000
-O+Dig24BxOFwjJEN1oBbk/VW/tA=	HST	1634712000
```

In diesem Beispiel meldet sich zunächst der Nutzer `Tester` am Standort `TST` an und wird im folgenden nur noch durch den Hash `HjLV...` referenziert.
In der Praxis wäre die erste Zeile aufgrund längerer Namen und einer vollständigen Adresseingabe deutlich länger.

Anschließend meldet sich der Nutzer `Klaus` am Standort `HST` an.
Darauf melden sich beide Nutzer nacheinander wieder ab.

### Writer

Der Writer enthält in sich alle bereits in das aktuelle Journal geschriebene Nutzer.
Wenn ein unbekannter Nutzer gemeldet wird, so muss eine `*`-Zeile erzeugt werden.
Weiterhin kennt der Writer den Ausgabeordner und einen offenen Ausgabe-Stream.
Zusätzlich enthält der Writer ein Mutex, um diesen Ausgabe-Stream Thread-sicher zu machen.

Wenn ein neuer Writer erstellt wird, wird zunächst geprüft, ob bereits eine Journal-Datei existiert.
Wenn dies der Fall ist, wird die entsprechende Datei eingelesen damit die bereits bekannten Nutzer geladen werden.

Anschließend wird der Ausgabe-Stream auf die Datei gesetzt.

Mit den Methoden `WriteEventUser` bzw. `WriteEventUserHash` können neue Ereignisse geschrieben werden.
Dabei existiert die zweite Methode, um einen vereinfachten Zugriff aus dem Programm heraus zuzulassen,
falls nur der Hash des Nutzers bekannt ist.

Weiterhin sollte die Methode `TrackJournalRotation` in einer Subroutine ausgeführt werden.
Diese Methode sorgt dafür, dass aller 24 Stunden auch bei laufendem Server eine neue Datei angelegt wird.

## Argp

Für das Einlesen der Argumente im Analyzer wurde zunächst versucht das Go-Standard-Modul `flag` einzusetzen.
Jedoch stellte sich schnell heraus, dass besonders die Arbeit mit *positional arguments* eher unschön ist und auch eine für den Nutzer schwer nachvollziehbare Auswertung bewirkt.
Im folgenden Aufruf des Programmes `xyz` würde der `flag`-Parser beide Argumente als *positional arguments* interpretieren, selbst wenn die Flag `--flag` ebenfalls existiert:

```
xyz Testdatei.txt --flag
```

Ebenso stellte sich die Arbeit mit Subcommands als sehr unschön heraus und so entstand der Gedanke ein entsprechendes eigenes Package zu schreiben.
Dieses Modul findet sich als `argp` mit dem in @fig:uml-argp dargestellten Klassenaufbau.

![UML-Klassendiagramm des `argp`-Packages](img/plantuml/pkg_argp.png){#fig:uml-argp}

Von der Grundidee folgt das Modul dem `flag`-Modul.
Verschiedene Flags werden über die Methoden für den jeweiligen Typ (z.B. `Bool` oder `Int`) zu einem `FlagSet` hinzugefügt.
Über die `FlagBuildArgs` kann die jeweilige Flag konfiguriert werden.
Dabei können verschiedene Namen angegeben werden, mit denen die Flag nutzbar sein soll.
Beide Schreibweisen mit `--` bzw. `-` sind beim Einlesen der Argumente erlaubt.
Zusätzlich kann ein Hilfstext angegeben werden und optional ein Text der als Standardwert in der Hilfe angezeigt wird.
Letzteres ist nützlich für Flags, deren Standardwert erst dynamisch bestimmt wird.

Als Rückgabewert wird von den Methoden ein Pointer zurückgegeben,
an dessen Speicheradresse beim erfolgreichen Parsen der finale Wert gespeichert wird.

Zusätzlich dazu können für jedes `FlagSet` noch positional Parameter definiert werden.
Diese werden analog zu den normalen Parametern mit den `Positional...`-Methoden definiert.

Weiterhin kann eine `SubcommandGroup` verwendet werden die verschiedene `Subcommand`s (`Subcommand` embedded `FlagSet`) gruppieren.
So ist es möglich eine Anwendung mit verschiedenen Unterbefehlen zu definieren.

Das Parsing kann abschließend durch einen Aufruf von `flagSet.ParseFlags(os.Args[1:])`,
beziehungsweise `subcommandGroup.ParseSubcommand(os.Args[1:])`, durchgeführt werden.
Im Fall der Subcommands wird zusätzlich zu einem möglichen Fehler noch der vom Nutzer verwendete `Subcommand` zurückgegeben.

Hilfetexte für alle Programme und Unterbefehle sind mittels der Flag `--help` anzeigbar.

## Token & QrCode {#sec:architecture-tokens}

Die Generation des Token nutzt das übergebene Kürzel des Standortes und die aktuelle Unix-Zeit, gerundet auf die Gültigkeitsdauer. 
Dies wird mit einem Doppelpunkt zu einem String von 16 Zeichen kombiniert, welcher an die Verschlüsselungsmethode übergeben wird.
Diese nutzt AES um den String, mit dem per Startparamter übergebenen Schlüssel, zu verschlüsseln. 
Dadurch hat man einen nicht lesbaren Token, welcher Standort spezifisch ist und zugleich die Zeit des Aufrufes enthällt.
Zur internen Nutzung ist desweiteren eine Methode zum entschlüsseln implementiert. 
Diese wird unteranderem von der Funktion Validate genutzt, um einen Token auf seine aktualität zu prüfen und den Standort zurückzugeben.

GetQrCode lässt zunächst einen Token erstellen und erstellt einen QrCode für den String aus der Startparameter URL und dem Token.
Dieser wird zurück gegeben als Byte Array.

## Webserver

### Aufbau

Um ein Webserver zu erstellen muss die Methode CreateWebserver aufgerufen werden.
Diese übernimmt als Parameter den `Port` als `uint` und ein `map` mit `Handlern`.
Die Rückgabewerte sind der startbereite `Server` und eine `destroy() Funktion` mit der
man den Server schließen kann.

Den Server ist mit RunWebserver(server) als https Server zu starten. 

RunWebservers baut den QRCodeWebserver und den LogIOServer auf. Hierfür sind 
zwei verschiedene Ports zu übergeben.

### Handling

#### QRCode Server

Der QrCode Server benutzt 3 Handler, wobei der `homeHandler` eine default Website erzeugt.
Im Weiteren gibt es den `qrPngHandler`, welcher für einen über den Get-Parameter location
übergebenen Ort ein temporären QRCode erzeugt. Diesen schreibt er in den ResponseWriter.
Der `qrPngHandler` wird vom `qrHandler` genutzt um den QRCode in das Template für die
QRCode Ausgabe einzubinden.

#### Login/Logout Server

Der LogIO Server besitzt auch 3 Handler. Hierbei finden im `loginHandler` das Login und 
`logoutHandler` das Logout statt. Dabei werden Token und Cookie auf Gekültigkeit überprüft,
bevor man sich an-/abmelden kann.

Der `cookieHandler` liest den Cookie des Nutzers aus (sofern vorhanden) und leitet wiefolgt 
an die anderen Handler weiter:

```
/ -> home
/?token=1234
    | cookie -> logged out = login.html login_data
    | cookie -> logged in  = logout.html
    | otherwise            = login.html
```

### Cookie {#sec:architecture-cookies}

Die Anmeldedaten eines Nutzers werden bei der ersten Anmeldung als Cookie
im Browser abgespeichert. Hierfür werden Name und Addresse zusammengefasst
und `:` getrennt erneut mit einem Secret gehashed abgespeichert

```
COOKIE=DATA:HASH
DATA=base64(NAME  \t  ADDRESS)
HASH=base64(hash(DATA  \t  SECRET))
```

Wenn sich ein Nutzer anmeldet, wird der Cookie, sofern vorhanden, ausgelesen
und ins in die Anmeldung voreingegeben. Hierbei wird überprüft, dass die Daten
mit den gehasheden Daten übereinstimmen. Das Secret welches zum hashen genutzt
wir über ein Flag beim Start übergeben. [#flags-webserver]


# Betriebsdokumentation

## Konfiguration der Orte {#sec:usage-locations}

Die Orte werden aus einer XML Datein übernommen, welche nach dem folgenden Schema aufgebaut ist.

```XML
<locations>
<location name="Mosbach" code="MOS"></location>
<location name="Bad Mergentheim" code="MGH"></location>
</locations>
```

Um einen neuen Ort hinzuzufügen kann lediglich in der XML Datei ein neuer location Tag, mit dem Namen und einem eindeutigem dreistelligem Code, eingefügt werden.

## Analyzer

Der Analyzer implementiert die geforderten CLI-Funktionalitäten zur Arbeit mit den Journals.

Die genaue Syntax für den Analyzer lässt sich mit der Flag `--help` (`-h`) oder dem entsprechenden Subcommand `help` anzeigen.
Im Folgenden wird die grobe Arbeitsweise mit dem Analyzer vorgestellt.

### Allgemeine Flags

#### Ortsdaten {#sec:usage-analyzer-general-locations}

Für fast alle Subcommands muss die Ortsliste geladen werden.
Hierfür kann mit der Flag `-l <Datei>` bzw. `--locations <Datei>` der Dateipfad zu der entsprechenden XML-Datei angegeben werden.
Das Format dieser Datei ist in [@sec:usage-locations] beschrieben.
Wenn die Flag nicht angegeben wird, dann wird standardmäßig `locations.xml` im aktuellen Ordner geladen.

#### Filtern von Personen {#sec:usage-analyzer-general-filters}

In einigen Subcommands kann nach ein oder mehreren Personen gefiltert werden.
Für diese Filterung kann nach Name (`-n <Name>` bzw. `--name <Name>`) und/oder Adresse (`-a <Adresse>` bzw. `--address <Adresse>`) gefiltert werden.

Beim Filtern wird die Groß- und Kleinschreibung ignoriert. Wenn beide Filter angegeben werden, wird eine UND-Verknüpfung angenommen.

### Anzeigen der Aufenthalte einer Person (`show-person`)

Mit dem Subcommand **`show-person`** kann eine Aufenthaltsliste für eine Person angegeben werden.
Für die Auswahl einer Person stehen die Filter aus [@sec:usage-analyzer-general-filters] zur Verfügung.
Als *positional argument* ist der Pfad zum untersuchenden Journal anzugegeben.

Beim Ausführen wird eine Liste der Orte und Uhrzeiten für die Person angegeben.

Beispiel-Befehl zum Anzeigen der Person `Tester` in der Datei `example.txt`:

```sh
lets-goooo-analyzer show-person example.txt --name Tester
```

### CSV-Export (`export`)

Anwesenheitslisten können mit dem Subcommand **`export`** erzeugt werden.

Als *positional argument* wird der Pfad zur Journal-Datei angegeben.

Mit der Flag `--output` bzw. `-o` kann eine Ausgabedatei spezifiziert werden.
Falls keine Ausgabedatei gesetzt ist wird der Pfad der Ausgabedatei aus dem Journal-Pfad abgeleitet.

Zusätzlich kann mit der Flag `--location` bzw. `--loc` nach einem Ort gefiltert werden.
Die Angabe kann hierbei durch den Anzeigenamen oder durch den internen Ortscode erfolgen.

Beispiel-Befehl zum Exportieren aller Kontakte am Ort `Mosbach` aus `example.txt`:

```sh
lets-goooo-analyzer export example.txt -o export.csv --loc Mosbach
```

### Erstellung von Kontaktlisten (`view-contacts`)

Mit dem Subcommand **`view-contacts`** können Kontaktlisten generiert werden.

Die zu verwendende Journal-Datei wird als *positional argument* angegeben.

Für die Auswahl einer Person können die in [@sec:usage-analyzer-general-filters] beschriebenen Filter eingesetzt werden.

Die **Ausgabe** erfolgt standardmäßig in einem gut lesbarem Format direkt auf der Konsole.
Mit der Flag `--output` bzw. `-o` kann eine Ausgabedatei angegeben werden.
Außerdem kann mit der Flag `--csv` die Ausgabe auf CSV umgestellt werden.

Beispiel-Befehl zum CSV-Export einer Kontaktliste für die Person `Tester` aus der Datei `example.txt`:

```sh
lets-goooo-analyzer view-contacts example.txt --name Tester --csv --csv-headers -o tester-contacts.csv
```

## Flags des Webservers {#sec:flags-webserver}

Auch der Webserver lässt mich mittels Flags stark konfigurieren.
Im Folgenden sind auch hierfür die wichtigsten Flags erläutert.

### Ortsdaten

Analog zum Analyzer in [@sec:usage-analyzer-general-locations].

### Web-Server

Die Ports für die Web-Server können mit `--frontend-port` bzw. `--backend-port` festgelegt werden.
Dabei läuft der Frontend-Server standardweise auf Port 4443 und das Backend auf Port 443.

Die entsprechenden TLS-Zertifikate lassen sich mit `--cert-file` und `--key-file` angeben,
standardmäßig werden die mitgelieferten selbst ausgestellten Zertifikate verwendet (`certification`-Ordner).

Um die QR-Codes generieren zu können, muss das Backend die öffentliche URL des Frontends kennen.
Diese kann mit `--frontend-base-url` gesetzt werden und zeigt standardmäßig auf den entsprechenden `localhost`-Port.

Weiterhin kann das Secret für [@sec:architecture-cookies] mit `--cookie-secret` gesetzt werden.
Andernfalls wird ein zufälliges Secret beim Start generiert.

### Tokens

Die Gültigkeitsdauer der Tokens kann mit `--token-valid-time` festgelegt werden.
Standard sind hier 120 Sekunden.

Auch für die Tokens wird ein Secret benötigt (siehe [@sec:architecture-tokens]).
Dieses kann mit `--token-secret` gesetzt werden und wird sonst zufällig bestimmt.

### Journals

Der Ablageort der Journal-Dateien kann mit dem Argument `--journals` festgelegt werden.
Weiterhin kann mit `--journal-file-permissions` die entsprechende Berechtigungsmaske für Unix-Systeme gesetzt werden.

# Anwenderdokumentation

# Mitgliedsbeitragsdokumentationen

## 1103207

## 3106335

## 4485500
