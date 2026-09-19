# Macchina di Turing

*[English version](README.en.md)*

Simulatore didattico di macchina di Turing, con nastro infinito, esecuzione
passo passo, tavola delle regole e grafo degli stati. Applicazione desktop
scritta in Go con [Wails](https://wails.io): il motore è Go, l'interfaccia è
una pagina web senza alcuna dipendenza esterna, il tutto in un unico
eseguibile.

![Schermata del simulatore](docs/schermata.png)

## Cos'è una macchina di Turing

È il modello di calcolo che Alan Turing descrisse nel 1936, e che ancora oggi
definisce cosa significa "calcolabile". La sua forza sta nell'essere la cosa
più semplice che possa ancora calcolare tutto ciò che è calcolabile. Gli
ingredienti sono quattro, e li ritrovi tutti nell'interfaccia:

- **Il nastro**, diviso in celle, ognuna contenente un simbolo. È
  potenzialmente infinito: qui si estende in entrambe le direzioni, e ogni
  cella mai scritta contiene il simbolo bianco `_`.
- **La testina**, che legge la cella sotto di sé, può riscriverla e si sposta
  di una posizione a sinistra o a destra, oppure resta ferma.
- **Il registro di stato**, che ricorda in quale stato si trova la macchina.
  Gli stati non sono memoria: sono le "situazioni" in cui la macchina può
  trovarsi, poche e con un nome.
- **La tavola di istruzioni**, che per ogni coppia (stato, simbolo letto) dice
  cosa scrivere, dove muoversi e in quale stato passare.

Non c'è altro. Nessuna variabile, nessun numero, nessuna aritmetica: sommare
due numeri significa spostare una testina avanti e indietro riscrivendo
simboli. Vederlo accadere è il modo migliore per capire perché il risultato di
Turing è così sorprendente.

## Il formato dei programmi

Un programma è un elenco di **quintuple**, una per riga, nella forma:

```
stato  letto  scritto  mossa  nuovo_stato
```

Il movimento è `L` (sinistra), `R` (destra) o `N` (fermo); si accettano anche
`S`, `D` e `F` all'italiana. Le righe che iniziano con `#` sono commenti, e due
direttive dichiarano lo stato iniziale e quelli finali:

```
inizio: q0
finali: qf
```

La macchina si ferma quando entra in uno stato finale, oppure quando per la
coppia (stato, simbolo letto) non esiste nessuna regola. Il programma deve
essere **deterministico**: due regole per la stessa coppia vengono rifiutate al
caricamento, con l'indicazione delle righe in conflitto.

### Esempio: incremento binario

Il programma caricato all'avvio somma 1 a un numero binario scritto sul nastro:

```
# Incremento binario
inizio: q0
finali: qf

q0 0 0 R q0      # scorre verso destra, lasciando intatte le cifre
q0 1 1 R q0
q0 _ _ L q1      # trovato il bianco: il numero è finito, torna indietro

q1 1 0 L q1      # 1 più riporto fa 0, e il riporto prosegue
q1 0 1 N qf      # 0 più riporto fa 1, e il riporto si assorbe: finito
q1 _ 1 N qf      # era tutto uni: scrive un 1 nuovo in testa
```

Lo stato `q0` ha un solo compito, raggiungere la fine del numero; `q1`
rappresenta "ho un riporto da sistemare". Provalo con `1011`, che diventa
`1100`, e poi con `111`: il riporto attraversa tutto il numero ed esce a
sinistra, scrivendo nella cella `-1`. Funziona senza trucchi solo perché il
nastro è infinito in entrambe le direzioni.

## Le due viste

Le stesse regole sono mostrate in due modi complementari. La **tavola** le
elenca come le hai scritte, ed è la vista testuale. Il **grafo degli stati** le
mostra come struttura: ogni stato è un nodo, ogni transizione una freccia
etichettata `(R;W;M)`, cioè simbolo letto, simbolo scritto, movimento. Il
cappio su `q0` racconta in un colpo d'occhio che quello stato "corre verso
destra finché legge cifre".

In entrambe le viste sono evidenziati lo stato corrente e la regola che sta per
scattare: prima di premere `Passo`, prova a indovinare quale si accenderà.

Oltre i dodici stati il grafo smette di essere leggibile, e la vista passa
automaticamente al solo registro di stato: un nodo grande che cambia etichetta
mentre la macchina cammina.

## Comandi

| Comando | Tastiera | Effetto |
|---|---|---|
| Passo | → | Esegue una singola transizione |
| Indietro | ← | Disfa l'ultimo passo |
| Avvia / Pausa | spazio | Esecuzione animata, con velocità regolabile |
| Corri fino in fondo | | Esegue fino all'arresto senza animazione |
| Azzera | | Riporta il nastro alla configurazione iniziale |

Il pulsante *Indietro* funziona perché ogni passo lascia un resoconto di com'era
la configurazione prima: disfarlo è quindi sempre possibile, a qualsiasi
profondità.

## Compilazione

Servono [Go](https://go.dev) 1.25 o superiore e la riga di comando di Wails v2:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
```

Poi, dalla radice del progetto:

```bash
wails dev      # esecuzione con ricarica automatica del frontend
wails build    # eseguibile in build/bin
go test ./tm/  # test del motore
```

Su macOS servono gli strumenti da riga di comando di Xcode
(`xcode-select --install`). Il frontend è HTML, CSS e JavaScript senza
framework, quindi non c'è nulla da costruire con npm.

## Architettura

Il progetto è diviso in due metà che non si conoscono.

Il package `tm` è il motore: nastro, regole, esecuzione, resoconti dei passi e
struttura del grafo. Non fa input né output, non sa che esiste una finestra, ed
è coperto dai test. Il nastro infinito è realizzato con due slice, una per le
celle non negative e una per quelle negative, che crescono solo quando la
testina ci scrive davvero.

Il resto è l'applicazione. `app.go` è il collegamento: espone al frontend i
metodi `Load`, `Step`, `Back`, `Run` e `Reset`, ognuno dei quali restituisce
una *fotografia* di ciò che va mostrato. Il frontend, in `frontend/dist`, non
contiene alcuna logica di macchina: chiama un metodo e ridisegna ciò che
riceve.

La separazione è netta per scelta: ciò che non cambia durante l'esecuzione
(tavola delle regole, grafo, posizioni dei nodi) viaggia una volta sola al
caricamento; a ogni passo viaggia solo ciò che cambia.

## Licenza

MIT
