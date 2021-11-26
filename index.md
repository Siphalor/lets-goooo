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

```java
// TODO
```

## Token

Die Generation des Token

## QrCode

# Anwenderdokumentation

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

# Betriebsdokumentation

# Mitgliedsbeitragsdokumentationen

## 1103207

## 3106335

## 4485500
