import { emitter } from "./modules/event-emitter.js";
import { HistoryService } from "./modules/data-service.js";

const state = {
  activeAccount: null,
  transactions: [],
  pagination: null,
  currentPage: 1
};

function showView(view) {
  document.getElementById("search-view").style.display = view === "search" ? "grid" : "none";
  document.getElementById("history-view").style.display = view === "history" ? "block" : "none";
}

// Re-usable refresh function that pulls current manual filter values
const refreshData = (page = state.currentPage) => {
  const pageSize = document.getElementById("pagesize-input").value;
  const sort = document.getElementById("sort-select").value;

  if (state.activeAccount) {
  HistoryService.fetchHistory(state.activeAccount, page, pageSize, sort);
}
};

const init = () => {
     const searchBtn = document.getElementById("btn-main-search");
     const reloadBtn = document.getElementById("btn-reload");
     const backBtn = document.getElementById("btn-back");
     const mainInput = document.getElementById("main-search-input");
     const sortSelect = document.getElementById("sort-select")
     // Initial Search
     searchBtn.onclick = () => {
        state.activeAccount = mainInput.value.trim();
        state.currentPage = 1;
        refreshData(1);
     }
     // Manual Reload / Apply Filters
     reloadBtn.onclick = () => {
        state.currentPage = 1; // Reset to page 1 when changing filters
        refreshData(1);
     }
     // Sort Dropdown auto-reload
     sortSelect.onchange = () => refreshData(1)
     backBtn.onclick = () => showView("search");
};

emitter.on("history:loaded", (payload) => {
     state.transactions = payload.history || [];
     state.pagination = payload["@metadata"];
     state.activeAccount = payload.accountNumber;
     state.currentPage = payload["@metadata"].current_page;
    
     showView("history");
     renderTable();
     renderPagination();
});

function renderTable() {
    const tbody = document.getElementById("history-tbody");
    document.getElementById("active-account-label").textContent = `History for ${state.activeAccount}`;

     tbody.innerHTML = state.transactions.map(t => `
          <tr>
               <td>${new Date(t.date).toLocaleDateString()}</td>
               <td><strong>${t.transaction_details}</strong></td>
               <td>${t.recipient_id.Valid ? t.recipient_id.Int64 : '—'}</td>
               <td class="text-right ${t.amount < 0 ? 'text-danger' : 'text-success'}">
                  ${t.amount < 0 ? '-' : '+'}$${Math.abs(t.amount).toLocaleString(undefined, {minimumFractionDigits: 2})}
               </td>
          </tr>
     `).join("");
}

function renderPagination() {
    const container = document.getElementById("pagination-controls");
    if (!state.pagination) return;

    const { current_page, last_page } = state.pagination;

    container.innerHTML = `
      <button class="page-btn" ${current_page === 1 ? 'disabled' : ''} id="prev-pg">Prev</button>
        <span>Page ${current_page} of ${last_page}</span>
      <button class="page-btn" ${current_page === last_page ? 'disabled' : ''} id="next-pg">Next</button>
    `;

  document.getElementById("prev-pg").onclick = () => refreshData(current_page - 1);
  document.getElementById("next-pg").onclick = () => refreshData(current_page + 1);
}

init();