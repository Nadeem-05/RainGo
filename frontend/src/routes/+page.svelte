<script lang="ts">
  import {
    Navbar,
    NavBrand,
    NavLi,
    NavUl,
    NavHamburger,
  } from "flowbite-svelte";
  import { onMount } from "svelte";
  import Chart from "chart.js/auto";

  onMount(() => {
    if (typeof document !== "undefined") {
      const data = [
        { year: 2010, count: 10 },
        { year: 2011, count: 20 },
        { year: 2012, count: 15 },
        { year: 2013, count: 25 },
        { year: 2014, count: 22 },
        { year: 2015, count: 30 },
        { year: 2016, count: 28 },
      ];

      const ctx = document.getElementById("acquisitions") as HTMLCanvasElement;
      new Chart(ctx, {
        type: "line",
        data: {
          labels: data.map((row) => row.year),
          datasets: [
            {
              label: "Acquisitions by year",
              data: data.map((row) => row.count),
              backgroundColor: "rgba(54, 162, 235, 0.2)",
              borderColor: "rgba(54, 162, 235, 1)",
              borderWidth: 2,
              pointBackgroundColor: "rgba(54, 162, 235, 1)",
              pointBorderColor: "#fff",
              pointRadius: 5,
              tension: 0.3, // Smoothens the line curve
            },
          ],
        },
        options: {
          responsive: true,
          plugins: {
            legend: {
              display: true,
              position: "top",
              labels: {
                color: "#fff",
              },
            },
          },
          scales: {
            x: {
              ticks: { color: "#fff" },
            },
            y: {
              ticks: { color: "#fff" },
              beginAtZero: true,
            },
          },
        },
      });
    }
  });

  const cards = [
    {
      title: "Try it out!",
      description: `
        <div class="flex flex-col space-y-6">
          <input type="text" class="mt-5 px-4 py-3 rounded-lg bg-gray-800 text-white focus:outline-none focus:ring-2 focus:ring-blue-500 text-lg" placeholder="Enter hash below" />
          <button  class="bg-blue-500 hover:bg-blue-600 text-white font-medium py-3 px-6 rounded-lg text-lg">Submit</button>
        </div>
      `,
      href: "/checkhash",
    },
    {
      title: "Statistics >",
      description: `
      <div class="chart-container"><canvas id="acquisitions"></canvas></div>`,
      href: "/chart",
    },
    {
      title: "Bulk Add Passwords",
      description: `
        <div class="flex flex-col space-y-6">
          <input type="file" class="mt-5 block w-full text-lg text-slate-500 file:mr-4 file:py-3 file:px-6 file:rounded-full file:border-0 file:text-lg file:font-semibold file:bg-violet-50 file:text-violet-700 hover:file:bg-violet-100" />
 <button  class="bg-green-500 hover:bg-green-600 text-white font-medium py-3 px-6 rounded-lg text-lg">Upload</button>
        </div>
      `,
      href: "/bulkupload",
    },
    {
      title: "Table Data",
      description: `
        <div class="mt-5 overflow-x-auto">
          <table class="w-full border-collapse">
            <thead>
              <tr>
                <th class="p-3 border text-left text-lg">Header 1</th>
                <th class="p-3 border text-left text-lg">Header 2</th>
                <th class="p-3 border text-left text-lg">Header 3</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td class="p-3 border text-lg">Data 1</td>
                <td class="p-3 border text-lg">Data 2</td>
                <td class="p-3 border text-lg">Data 3</td>
              </tr>
              <tr>
                <td class="p-3 border text-lg">Data 4</td>
                <td class="p-3 border text-lg">Data 5</td>
                <td class="p-3 border text-lg">Data 6</td>
              </tr>
            </tbody>
          </table>
        </div>
      `,
      href: "/table",
    },
  ];
</script>

<div>
  <Navbar rounded class="bg-transparent dark">
    <NavBrand href="/">
      <span class="self-center text-xl font-semibold dark:text-white"
        >Raingo</span
      >
    </NavBrand>
    <NavHamburger />
    <NavUl>
      <NavLi href="/" activeClass="active" class="dark:text-white">Home</NavLi>
      <NavLi href="/checkhash" class="dark:text-white">Converter</NavLi>
      <NavLi href="/chart" class="dark:text-white">Stats</NavLi>
      <NavLi href="/bulkupload" class="dark:text-white">Upload</NavLi>
      <NavLi href="/table" class="dark:text-white">Table</NavLi>
    </NavUl>
  </Navbar>
</div>

<div class="grid-container">
  {#each cards as card}
    <div class="card text-white">
      <div class="card-content">
        <a href={card.href}>
          <h5 class="card-title">{card.title}</h5>

          {@html card.description}
        </a>
      </div>
    </div>
  {/each}
</div>

<style>
  :global(html) {
    background-image: linear-gradient(
      0deg,
      rgba(32, 42, 68, 1) 25%,
      rgba(9, 22, 46, 1) 50%,
      rgba(0, 1, 25, 1) 75%,
      rgba(0, 0, 0, 1) 100%
    );
    animation: slide 3s ease-in-out infinite alternate;
    background-attachment: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    z-index: -1;
    margin: auto auto;
  }

  :global(.app-container) {
    min-height: 100vh;
    position: relative;
    z-index: 1;
    overflow-x: hidden;
  }

  /* Main content container */
  .grid-container {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 2rem;
    max-width: 800px;
    margin: 0 auto;
    padding: 1rem;
    position: relative;
    margin: auto auto;
  }

  /* Mobile responsiveness */
  @media (max-width: 768px) {
    .grid-container {
      grid-template-columns: 1fr;
      padding: 1rem;
      padding-bottom: 2rem;
    }

    :global(.card) {
      width: 100%;
      margin-bottom: 1rem;
    }
  }

  .card {
    background-color: #1f2937;
    border-radius: 0.5rem;
    padding: 2.5rem;
    box-shadow:
      0 4px 6px -1px rgba(0, 0, 0, 0.1),
      0 2px 4px -1px rgba(0, 0, 0, 0.06);
    transition: box-shadow 0.3s ease;
  }

  .card:hover {
    box-shadow:
      0 10px 15px -3px rgba(0, 0, 0, 0.1),
      0 4px 6px -2px rgba(0, 0, 0, 0.05);
  }

  .chart-container {
    width: 100%;
    max-width: 100%;
    height: 300px;
    margin: 0 auto;
  }

  /* Fix for mobile scrolling */
  :global(body) {
    -webkit-overflow-scrolling: touch;
  }
</style>
