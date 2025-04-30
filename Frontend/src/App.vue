<!-- App.vue -->
<template>

    <div class="container">
        <h1>Fraud Detection Dashboard</h1>

        <div class="search-form">
            <input v-model="userId" placeholder="Enter User ID" class="input-field" />
            <button @click="fetchFlaggedTransactions" class="search-button">
                Search
            </button>
        </div>

        <div v-if="loading" class="loading">
            Loading...
        </div>

        <div v-if="error" class="error">
            {{ error }}
        </div>

        <div v-if="flaggedTransactions.length" class="results-container">
            <h2>Flagged Transactions</h2>
            <table class="transactions-table">
                <thead>
                    <tr>
                        <th>Transaction ID</th>
                        <th>Amount</th>
                        <th>Timestamp</th>
                        <th>Merchant</th>
                        <th>Location</th>
                        <th>Reason</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="transaction in paginatedTransactions" :key="transaction.TransactionID">
                        <td>{{ transaction.TransactionID }}</td>
                        <td>${{ transaction?.Amount }}</td>
                        <td>{{ formatDate(transaction.Timestamp) }}</td>
                        <td>{{ transaction.Merchant }}</td>
                        <td>{{ transaction.Location }}</td>
                        <td>{{ transaction.Reason }}</td>
                    </tr>
                </tbody>
            </table>
            <div class="pagination-controls" v-if="flaggedTransactions.length > pageSize">
                <button @click="prevPage" :disabled="currentPage === 1">Previous</button>
                <span>Page {{ currentPage }} of {{ totalPages }}</span>
                <button @click="nextPage" :disabled="currentPage === totalPages">Next</button>
            </div>

        </div>

        <div v-if="flaggedTransactions.length" class="map-container">
            <h2>Transaction Map</h2>
            <div id="map" style="height: 500px;"></div>
        </div>
    </div>

</template>

<script>
import axios from 'axios';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';

export default {
    name: 'App',
    data() {
        return {
            userId: '',
            flaggedTransactions: [],
            loading: false,
            error: null,
            map: null,
            markers: [],
            currentPage: 1,
            pageSize: 10, // Show 5 transactions per page
        };
    },
    computed: {
        paginatedTransactions() {
            const start = (this.currentPage - 1) * this.pageSize;
            const end = start + this.pageSize;
            return this.flaggedTransactions.slice(start, end);
        },
        totalPages() {
            return Math.ceil(this.flaggedTransactions.length / this.pageSize);
        }
    },
    methods: {
        async fetchFlaggedTransactions() {
            if (!this.userId) {
                this.error = 'Please enter a User ID';
                return;
            }

            this.loading = true;
            this.error = null;

            try {
                const response = await axios.get(`http://localhost:6080/api/fraud-check?userId=${this.userId}`);

                // Adjust your API field mapping here
                this.flaggedTransactions = response.data.map(tx => ({
                    TransactionID: tx.transactionId,
                    Amount: tx.amount,
                    Timestamp: tx.timestamp,
                    Merchant: tx.merchant,
                    Location: tx.location,
                    Reason: tx.reason,
                }));

                console.log(this.flaggedTransactions);

                this.currentPage = 1; // Reset to first page

                this.$nextTick(() => {
                    this.initMap();
                });
            } catch (error) {
                this.error = `Error fetching data: ${error.message}`;
                this.flaggedTransactions = [];
            } finally {
                this.loading = false;
            }
        },

        initMap() {
            if (this.flaggedTransactions.length === 0) return;

            if (this.map) {
                this.map.remove();
            }

            this.map = L.map('map').setView([0, 0], 2);

            L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
                attribution: '&copy; OpenStreetMap contributors'
            }).addTo(this.map);

            this.markers = [];

            this.flaggedTransactions.forEach(transaction => {
                const locationParts = transaction.Location.split(',');
                if (locationParts.length === 2) {
                    const lat = parseFloat(locationParts[0]);
                    const lng = parseFloat(locationParts[1]);

                    if (!isNaN(lat) && !isNaN(lng)) {
                        const marker = L.marker([lat, lng])
                            .addTo(this.map)
                            .bindPopup(`
                                <b>Transaction ID:</b> ${transaction.TransactionID}<br>
                                <b>Amount:</b> $${transaction.Amount.toFixed(2)}<br>
                                <b>Reason:</b> ${transaction.Reason}
                            `);

                        this.markers.push(marker);
                    }
                }
            });

            if (this.markers.length > 0) {
                const group = new L.featureGroup(this.markers);
                this.map.fitBounds(group.getBounds().pad(0.1));
            }
        },

        formatDate(dateString) {
            const date = new Date(dateString);
            return date.toLocaleString();
        },

        nextPage() {
            if (this.currentPage < this.totalPages) {
                this.currentPage++;
            }
        },

        prevPage() {
            if (this.currentPage > 1) {
                this.currentPage--;
            }
        }
    }
};
</script>

<style>
.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
}

h1 {
    text-align: center;
    margin-bottom: 30px;
}

.search-form {
    display: flex;
    margin-bottom: 20px;
}

.input-field {
    flex: 1;
    padding: 10px;
    font-size: 16px;
    border: 1px solid #ddd;
    border-radius: 4px 0 0 4px;
}

.search-button {
    padding: 10px 20px;
    background-color: #4CAF50;
    color: white;
    border: none;
    border-radius: 0 4px 4px 0;
    cursor: pointer;
    font-size: 16px;
}

.search-button:hover {
    background-color: #45a049;
}

.loading {
    text-align: center;
    padding: 20px;
    font-size: 18px;
}

.error {
    background-color: #ffeeee;
    color: #ff0000;
    padding: 10px;
    border-radius: 4px;
    margin-bottom: 20px;
}

.results-container {
    margin-bottom: 30px;
}

.transactions-table {
    width: 100%;
    border-collapse: collapse;
    margin-top: 10px;
}

.transactions-table th,
.transactions-table td {
    padding: 10px;
    border: 1px solid #ddd;
    text-align: left;
}

.transactions-table th {
    background-color: #f2f2f2;
}

.transactions-table tr:hover {
    background-color: #f5f5f5;
}

.map-container {
    margin-top: 30px;
}

#map {
    border: 1px solid #ddd;
    border-radius: 4px;
}

.pagination-controls {
    margin-top: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
}

.pagination-controls button {
    padding: 8px 12px;
    background-color: #4CAF50;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
}

.pagination-controls button:disabled {
    background-color: #ccc;
    cursor: not-allowed;
}

</style>