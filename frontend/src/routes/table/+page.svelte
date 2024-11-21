<script lang="ts">
  import { A, Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell } from 'flowbite-svelte';
  import {
    Navbar,
    NavBrand,
    NavLi,
    NavUl,
    NavHamburger,
  } from "flowbite-svelte";

  import type { main } from "$lib/wailsjs/go/models";
  import { GetEntries } from "$lib/wailsjs/go/main/App";
  //@ts-ignore
  import { GetTotalEntries } from "$lib/wailsjs/go/main/App";
  import { onMount } from "svelte";

  let entries: main.Entry[] = [];
  let loading = true;
  let totalItems = 0; // Initialize to 0
  let itemsPerPage = 10;
  let usePostgres = false; // Toggle for database selection
  $: totalPages = Math.ceil(totalItems / itemsPerPage); // Make it reactive
  let error: string | null = null;
  let currentPage: number = 1;

  $: helper = { 
    start: (currentPage - 1) * itemsPerPage + 1,
    end: Math.min(currentPage * itemsPerPage, totalItems),
    total: totalItems
  };

  async function fetchData(page: number) {
    loading = true;
    error = null;
    try {
        const result = await GetEntries(page, usePostgres); // Pass `usePostgres` to the backend
        if (result) {
            entries = result;
        }
    } catch (err) {
        console.error("Error fetching entries:", err);
        error = err instanceof Error ? err.message : 'An error occurred while fetching data';
    } finally {
        loading = false;
    }
  }

  async function getTotalItems() {
    try {
        const total = await GetTotalEntries(usePostgres); 
        totalItems = total;
    } catch (err) {
        console.error("Error fetching total entries:", err);
        error = err instanceof Error ? err.message : 'An error occurred while fetching total count';
    }
  }

  onMount(async () => {
    await getTotalItems();
    await fetchData(currentPage);
  });

  const previous = async (): Promise<void> => {
    if (currentPage > 1) {
      currentPage -= 1;
      await fetchData(currentPage);
    }
  };

  const next = async (): Promise<void> => {
    if (currentPage < totalPages) {
      currentPage += 1;
      await fetchData(currentPage);
    }
  };

  const toggleDatabase = async (): Promise<void> => {
    usePostgres = !usePostgres; // Toggle between databases
    await getTotalItems(); // Recalculate total items for the selected database
    await fetchData(currentPage); // Fetch data from the selected database
  };
</script>

<Navbar rounded class="bg-transparent dark shadow-md">
  <NavBrand href="/">
      <img src="src/images/favicon.png" class="me-3 h-6 sm:h-9" alt="Flowbite Logo" />
      <span class="self-center text-xl font-semibold dark:text-white">Flowbite</span>
  </NavBrand>
  <NavHamburger />
  <NavUl>
      <NavLi href="/" activeClass="active" class="dark:text-white">Home</NavLi>
      <NavLi href="/about" class="dark:text-white">About</NavLi>
      <NavLi href="/docs/components/navbar" class="dark:text-white">Navbar</NavLi>
      <NavLi href="/pricing" class="dark:text-white">Pricing</NavLi>
      <NavLi href="/contact" class="dark:text-white">Contact</NavLi>
  </NavUl>
</Navbar>

<div class="container mx-auto">
  <button 
    on:click={toggleDatabase} 
    class="mb-4 px-4 py-2 text-sm font-medium -mt-48 border rounded-lg bg-gray-800 border-gray-700 text-gray-400  hover:bg-gray-700"
  >
    {usePostgres ? "Switch to Local DB" : "Switch to PostgreSQL"}
  </button>

  {#if loading}
    <div class="text-center p-4 dark:text-white">Loading...</div>
  {:else if error}
    <div class="text-center p-4 text-red-500">{error}</div>
  {:else}
    <Table classInput="dark float-end1" classSvgDiv="hidden" placeholder="Search by maker name" hoverable={true} class="dark w-full  mx-auto mb-5 shadow-md rounded-lg">
      <TableHead>
          <TableHeadCell>ID</TableHeadCell>
          <TableHeadCell>Maker</TableHeadCell>
          <TableHeadCell>Hash</TableHeadCell>
          <TableHeadCell>Type</TableHeadCell>
      </TableHead>
      <TableBody tableBodyClass="divide-y">
        {#each entries as item}
          <TableBodyRow>
            <TableBodyCell>{item.id}</TableBodyCell>
            <TableBodyCell>{item.pwd}</TableBodyCell>
            <TableBodyCell>{item.hash}</TableBodyCell>
            <TableBodyCell>{item.type}</TableBodyCell>
          </TableBodyRow>  
        {/each}
      </TableBody>
    </Table>
  {/if}
</div>

<div class="flex flex-col items-center dark justify-center gap-4 -mt-5">
  <div class="text-sm text-gray-700 dark:text-gray-400">
    Showing <span class="font-semibold text-gray-900 dark:text-white">{helper.start}</span>
    to
    <span class="font-semibold text-gray-900 dark:text-white">{helper.end}</span>
    of
    <span class="font-semibold text-gray-900 dark:text-white">{helper.total}</span>
    Entries
  </div>
  <div class="flex justify-center gap-2">
    <button 
      on:click={previous}
      disabled={currentPage === 1}
      class="px-4 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-lg dark:bg-gray-800 dark:border-gray-700 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
    >
      Previous
    </button>
    <button 
      on:click={next}
      disabled={currentPage === totalPages}
      class="px-4 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-lg dark:bg-gray-800 dark:border-gray-700 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
    >
      Next
    </button>
  </div>
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
      background-attachment: fixed;
      overflow-y: scroll;
      height: 100%;
      width: 100%;
      margin: 0;
      animation: slide 3s ease-in-out infinite alternate;
  }

  .container {
      padding: 1rem;
  }

  .dark {
      color: white;
  }

  .shadow-md {
      box-shadow: 0px 4px 8px rgba(0, 0, 0, 0.2);
  }

  .rounded-lg {
      border-radius: 0.5rem;
  }
</style>

