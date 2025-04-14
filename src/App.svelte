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
    chartData["MFI"] = chartData["MFI"].map((value) => {
      return value == 0.0 ? null : value;
    });
    chartData["StochK"] = chartData["StochK"].map((value) => {
      return value == 0.0 ? null : value;
    });
    chartData["StochD"] = chartData["StochD"].map((value) => {
      return value == 0.0 ? null : value;
    });
    chartData["BBH"] = chartData["BBH"].map((value) => {
      return value == 0.0 ? null : value;
    });
    chartData["BBM"] = chartData["BBM"].map((value) => {
      return value == 0.0 ? null : value;
    });
    chartData["BBL"] = chartData["BBL"].map((value) => {
      return value == 0.0 ? null : value;
    });
    chartData["SMA"] = chartData["SMA"].map((value) => {
      return value == 0.0 ? null : value;
    });
    chartData["ADX"] = chartData["ADX"].map((value) => {
      return value == 0.0 ? null : value;
    });
    await updateChart();
  }

  async function initChart() {
    let options = {
      animation: false,
      xAxis: [
        { type: "category", gridIndex: 0 },
        { type: "category", gridIndex: 1 },
        { type: "category", gridIndex: 2 },
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
        },
        {
          type: "value",
          scale: true,
          gridIndex: 2,
          splitLine: { show: false },
        },
      ],

      tooltip: {
        trigger: "axis",
      },
      dataZoom: [
        {
          type: "slider",
          start: 90,
          end: 100,
          xAxisIndex: [0, 1, 2],
        },
        {
          type: "inside",
          xAxisIndex: [0, 1, 2],
        },
      ],
      grid: [
        {
          left: "10%",
          right: "10%",
          bottom: 240,
        },
        {
          left: "10%",
          right: "10%",
          height: 60,
          bottom: 70,
        },
        {
          left: "10%",
          right: "10%",
          height: 60,
          bottom: 160,
        },
      ],
    };
    chart.setOption(options);
  }

  async function updateChart() {
    let options = {
      title: {
        text: symbol,
        left: "center",
      },
      dataset: {
        source: chartData,
      },
      series: [
        {
          type: "candlestick",
          itemStyle: {
            color: "#47b262",
            color0: "#eb5454",
            borderColor: "#47b262",
            borderColor0: "#eb5454",
          },
          name: "Close",
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
            silent: true,
            symbol: ["none", "none"],
            data: [{ yAxis: chartData["Close"].slice(-1) }],
          },
          data: [],
          tooltip: {
            valueFormatter: (value) => (value ? value.toFixed(2) : ""),
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
            silent: true,
            symbol: ["none", "none"],
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
          color: "red",
          lineStyle: {
            width: 1,
          },
          tooltip: {
            valueFormatter: (value) => (value ? value.toFixed(2) : ""),
          },
        },
        {
          type: "line",
          name: "BBM",
          seriesLayoutBy: "column",
          encode: {
            x: "Timestamp",
            y: "BBM",
          },
          showSymbol: false,
          color: "blue",
          lineStyle: {
            width: 1,
          },
          tooltip: {
            valueFormatter: (value) => (value ? value.toFixed(2) : ""),
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
          color: "green",
          lineStyle: {
            width: 1,
          },
          tooltip: {
            valueFormatter: (value) => (value ? value.toFixed(2) : ""),
          },
        },
        {
          type: "line",
          name: "ADX",
          showSymbol: false,
          xAxisIndex: 1,
          yAxisIndex: 1,
          smooth: true,
          seriesLayoutBy: "column",
          encode: {
            x: "Timestamp",
            y: "ADX",
          },
          lineStyle: {
            width: 1,
          },
          tooltip: {
            valueFormatter: (value) => (value ? value.toFixed(2) : ""),
          },
        },
        {
          type: "line",
          xAxisIndex: 1,
          yAxisIndex: 1,
          markLine: {
            data: [{ yAxis: 25.0 }],
            silent: true,
            symbol: ["none", "none"],
            lineStyle: {
              color: "gray",
              width: 1,
            },
          },
          data: [],
        },
        {
          type: "line",
          name: "MFI",
          showSymbol: false,
          xAxisIndex: 2,
          yAxisIndex: 2,
          smooth: true,
          seriesLayoutBy: "column",
          encode: {
            x: "Timestamp",
            y: "MFI",
          },
          lineStyle: {
            width: 1,
          },
          tooltip: {
            valueFormatter: (value) => (value ? value.toFixed(2) : ""),
          },
        },
        {
          type: "line",
          xAxisIndex: 2,
          yAxisIndex: 2,
          markLine: {
            data: [{ yAxis: 70.0 }],
            silent: true,
            symbol: ["none", "none"],
            lineStyle: {
              color: "gray",
              width: 1,
            },
          },
          data: [],
        },
        {
          type: "line",
          xAxisIndex: 2,
          yAxisIndex: 2,
          markLine: {
            data: [{ yAxis: 30.0 }],
            silent: true,
            symbol: ["none", "none"],
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
    initChart();
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
