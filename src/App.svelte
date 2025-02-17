<script>
  import * as echarts from "echarts";
  import { onMount } from "svelte";

  const updateInterval = 10000;

  let chart;
  let tickers = $state([]);
  let symbol = $state(window.location.hash.replace("#", ""));
  let chartData = $state({});
  let timer;

  window.addEventListener("hashchange", () => {
    symbol = window.location.hash.replace("#", "");
  });

  window.addEventListener("resize", () => {
    chart.resize();
  });

  $effect(() => {
    if (symbol != "") {
      if (timer) {
        clearInterval(timer);
      }
      fetchChartData();
      timer = setInterval(fetchChartData, updateInterval);
      scroll(0, 0);
    }
  });

  async function fetchTickers() {
    const response = await fetch("/api/tickers/");
    tickers = await response.json();
    tickers.sort((a, b) => a.Symbol.localeCompare(b.Symbol));
  }

  async function fetchChartData() {
    const response = await fetch("/api/tickers/" + symbol);
    chartData = await response.json();
    await updateChart();
  }

  async function updateChart() {
    let options = {
      title: {
        text: symbol,
        left: "center",
      },
      animation: false,
      dataset: {
        source: chartData,
      },
      xAxis: [
        { type: "time", gridIndex: 0 },
        { type: "time", gridIndex: 1 },
      ],
      yAxis: [
        {
          type: "value",
          scale: true,
          gridIndex: 0,
          splitLine: { show: false },
        },
        {
          type: "value",
          scale: true,
          gridIndex: 1,
          splitLine: { show: false },
          min: 0,
          max: 100,
        },
      ],

      tooltip: {
        trigger: "axis",
      },
      dataZoom: [
        {
          type: "slider",
          start: 70,
          end: 100,
          xAxisIndex: [0, 1],
        },
        {
          type: "inside",
          xAxisIndex: [0, 1],
        },
      ],
      grid: [
        {
          left: "10%",
          right: "10%",
          bottom: 200,
        },
        {
          left: "10%",
          right: "10%",
          height: 80,
          bottom: 80,
        },
      ],
      series: [
        {
          type: "candlestick",
          itemStyle: {
            color: "#47b262",
            color0: "#eb5454",
            borderColor: "#47b262",
            borderColor0: "#eb5454",
          },
          name: "OHLC",
          encode: {
            x: "Timestamp",
            y: ["Open", "Close", "Low", "High"],
          },
          seriesLayoutBy: "column",
        },
        {
          type: "line",
          markLine: {
            lineStyle: {
              color: chartData["Close"].slice(-1) > chartData["Open"].slice(-1) ? "green" : "red",
              width: 1,
            },
            data: [{ yAxis: chartData["Close"].slice(-1) }],
          },
          data: [],
          tooltip: {
            valueFormatter: (value) => value.toFixed(2),
          },
        },
        {
          type: "line",
          markLine: {
            data: [{ yAxis: chartData["BuyPrice"].toFixed(2) }],
            lineStyle: {
              color: "blue",
              width: 1,
            },
          },
          data: [],
        },
        {
          type: "line",
          name: "BBH",
          seriesLayoutBy: "column",
          encode: {
            x: "Timestamp",
            y: "BBH",
          },
          showSymbol: false,
          color: "gray",
          lineStyle: {
            width: 1,
          },
          tooltip: {
            valueFormatter: (value) => value.toFixed(2),
          },
        },
        {
          type: "line",
          name: "BBL",
          seriesLayoutBy: "column",
          encode: {
            x: "Timestamp",
            y: "BBL",
          },
          showSymbol: false,
          color: "gray",
          lineStyle: {
            width: 1,
          },
          tooltip: {
            valueFormatter: (value) => value.toFixed(2),
          },
        },
        {
          type: "line",
          name: "StochK",
          data: chartData["stochk"],
          showSymbol: false,
          xAxisIndex: 1,
          yAxisIndex: 1,
          seriesLayoutBy: "column",
          encode: {
            x: "Timestamp",
            y: "StochK",
          },
          smooth: true,
          tooltip: {
            valueFormatter: (value) => value.toFixed(2),
          },
        },
        {
          type: "line",
          name: "StochD",
          data: chartData["stochd"],
          showSymbol: false,
          xAxisIndex: 1,
          yAxisIndex: 1,
          smooth: true,
          seriesLayoutBy: "column",
          encode: {
            x: "Timestamp",
            y: "StochD",
          },
          tooltip: {
            valueFormatter: (value) => value.toFixed(2),
          },
        },
        {
          type: "line",
          xAxisIndex: 1,
          yAxisIndex: 1,
          markLine: {
            data: [{ yAxis: 20.0 }, { yAxis: 80.0 }],
            lineStyle: {
              color: "gray",
              width: 1,
            },
          },
          data: [],
        },
      ],
    };
    chart.setOption(options);
  }

  onMount(async () => {
    await fetchTickers();
    setInterval(fetchTickers, updateInterval);
  });

  function charts(node) {
    chart = echarts.init(node, null, { renderer: "svg" });
  }
</script>

<main class="container">
  {#if tickers.length === 0}
    <p>Loading...</p>
  {/if}
  {#if symbol != ""}
    <div id="chart" use:charts></div>
  {/if}
  <table class="striped">
    <thead>
      <tr>
        <th>Symbol</th>
        <th class="right">Buy Price</th>
        <th class="right">Close</th>
        <th class="right">Change</th>
        <th class="center">Signal</th>
      </tr>
    </thead>
    <tbody>
      {#each tickers as ticker}
        <tr>
          <td><a href="#{ticker.Symbol}">{ticker.Symbol}</a></td>
          <td class="right"
            >{#if ticker.BuyPrice > 0}${ticker.BuyPrice.toFixed(2)}{/if}</td
          >
          <td class="right">${ticker.Close.toFixed(2)}</td>
          <td class="right"
            >{#if ticker.BuyPrice > 0}{ticker.Change.toFixed(2)}%{/if}</td
          >
          <td class="center">{ticker.Signal}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</main>

<style>
  #chart {
    width: auto;
    height: 600px;
  }
  .right {
    text-align: right;
  }
  .center {
    text-align: center;
  }
  a:link {
    text-decoration: none;
  }
  a:visited {
    text-decoration: none;
  }
  a:hover {
    text-decoration: none;
  }
  a:active {
    text-decoration: none;
  }
</style>
