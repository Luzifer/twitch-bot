new Vue({
  computed: {
    quoteItems() {
      return this.quotes.map((q, i) => ({ id: i + 1, quote: q }))
    },
  },

  data: {
    fields: [
      { key: 'id', label: 'ID', sortable: true, sortDirection: 'desc' },
      { key: 'quote' },
    ],

    quotes: [],
  },

  el: '#app',

  mounted() {
    axios.get(window.location.href)
      .then(res => {
        this.quotes = res.data
      })
  },
})
