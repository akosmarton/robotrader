<script>
  import * as echarts from "echarts";
  import { onMount } from "svelte";

  let chart = null;
  let tickers = $state([]);
  let symbol = $state(window.location.hash.replace("#", ""));
  let chartData = $state({});

  window.addEventListener("hashchange", () => {
    symbol = window.location.hash.replace("#", "");
  });

  window.addEventListener("resize", () => {
    chart.resize();
  });

  $effect(() => {
    if (symbol != "") {
      fetchChartData(symbol);
      scroll(0,0);
    }
  });

  async function fetchTickers() {
    const response = await fetch("/api/tickers/");
    tickers = await response.json();
    tickers.sort((a, b) => a.Symbol.localeCompare(b.Symbol));
  }

  async function fetchChartData(symbol) {
    const response = await fetch("/api/tickers/" + symbol);
    const data = await response.json();
    let ohlc = [];
    let bbh = [];
    let bbl = [];
    let stochk = [];
    let stochd = [];

    data.Timestamp.forEach((x, i) => {
      ohlc[i] = [x, data.Open[i], data.Close[i], data.Low[i], data.High[i]];
      bbh[i] = [x, data.BBH[i] ? data.BBH[i].toFixed(2) : null];
      bbl[i] = [x, data.BBL[i] ? data.BBL[i].toFixed(2) : null];
      stochk[i] = [x, data.StochK[i] ? data.StochK[i].toFixed(2) : null];
      stochd[i] = [x, data.StochD[i] ? data.StochD[i].toFixed(2) : null];
    });
    chartData = {
      ohlc,
      bbh,
      bbl,
      stochk,
      stochd,
      open: data.Open[data.Open.length - 1],
      close: data.Close[data.Close.length - 1],
      buyPrice: data.BuyPrice,
    };
    await updateChart();
  }

  async function updateChart() {
    let options = {
      title: {
        text: symbol,
        left: "center",
      },
      animation: false,
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
          data: chartData["ohlc"],
        },
        {
          type: "line",
          markLine: {
            lineStyle: {
              color: chartData["close"] > chartData["open"] ? "green" : "red",
              width: 1,
            },
            data: [{ yAxis: chartData["close"] }],
          },
        },
        {
          type: "line",
          markLine: {
            data: [{ yAxis: chartData["buyPrice"] }],
            lineStyle: {
              color: "blue",
              width: 1,
            },
          },
        },
        {
          type: "line",
          name: "BBH",
          data: chartData["bbh"],
          showSymbol: false,
          color: "gray",
          lineStyle: {
            width: 1,
          },
        },
        {
          type: "line",
          name: "BBL",
          data: chartData["bbl"],
          showSymbol: false,
          color: "gray",
          lineStyle: {
            width: 1,
          },
        },
        {
          type: "line",
          name: "StochK",
          data: chartData["stochk"],
          showSymbol: false,
          xAxisIndex: 1,
          yAxisIndex: 1,
          smooth: true,
        },
        {
          type: "line",
          name: "StochD",
          data: chartData["stochd"],
          showSymbol: false,
          xAxisIndex: 1,
          yAxisIndex: 1,
          smooth: true,
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
        },
      ],
    };
    chart.setOption(options);
  }

  onMount(async () => {
    await fetchTickers();
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
        <th>Buy Price</th>
        <th>Close</th>
        <th>Change</th>
        <th>Signal</th>
      </tr>
    </thead>
    <tbody>
      {#each tickers as ticker}
        <tr>
          <td><a href="#{ticker.Symbol}">{ticker.Symbol}</a></td>
          <td
            >{#if ticker.BuyPrice > 0}{ticker.BuyPrice}{/if}</td
          >
          <td>{ticker.Close}</td>
          <td
            >{#if ticker.BuyPrice > 0}{ticker.Change.toFixed(2)}%{/if}</td
          >
          <td>{ticker.Signal}</td>
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
</style>
