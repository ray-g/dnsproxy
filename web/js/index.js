var apiURL = window.location.origin + "/"

var app = new Vue({
  el: '#app',
  data: {
    qps: [],
    totalCached: 0,
    numDomainTotal: 0,
    numDomainBlocked: 0,
    queryTotal: 0,
    queryBlocked: 0,
    percentageBlocked: 0,
    timeStarted: 0,
    timeElapsed: 0,
    lastUpdated: 0,
    loading: true,
    loadingText: "loading data...",
    active: false,
    autoUpdate: false,
    autoUpdateId: 0,
    autoUpdateInterval: 10,
    domainQuestion: "",
    domainAnswerCache: "",
    domainAnswerQuery: "",
  },
  created: function () {
    this.fetchStats()
    this.getActive()
    this.pollActive()
  },
  methods: {
    queryDomain: function() {
      var self = this
      self.domainAnswerCache = ""
      self.domainAnswerQuery = ""
      $.get(apiURL + 'query/' + self.domainQuestion, function (data) {
        self.domainAnswerCache = data.cache
        if (data.query) {
          self.domainAnswerQuery = data.query
        } else {
          self.domainAnswerQuery = "didn't get a valid answer"
        }
      })
    },
    fetchStats: function () {
      var self = this
      $.get(apiURL + 'stats', function (data) {
        self.qps = data.stats.qps
        self.timeStarted = new Date(data.stats.time_started  * 1000)
        self.lastUpdated = data.stats.time_last
        self.numDomainTotal = data.stats.domain_count
        self.numDomainBlocked = data.stats.domain_blocked
        self.queryTotal = data.stats.query_count
        self.queryBlocked = data.stats.query_blocked
        if (self.queryTotal > 0) {
          self.percentageBlocked = self.queryBlocked/self.queryTotal*100
        }
        var now = new Date()
        self.timeElapsed = ( now - self.timeStarted ) / 1000
        self.generateStats()
      })
    },
    generateStats: function () {
      var self = this

      if (self.qps) {
        self.generateChart()
      } else {
        self.loading = false
      }
    },
    generateChart: function () {
      var self = this
      var cols = []
      var xPlot = ['x']
      var yPlot = ['qps']

      var timeStart = self.lastUpdated - self.qps.length
      var times = new Array(self.qps.length)

      for (i=0; i<self.qps.length; i++) {
        if (self.qps[i] <= 0) {
          continue
        }
        timestamp = timeStart + i
        times[i] = timestamp
        xPlot.push(timestamp)
        yPlot.push(self.qps[i])
      }

      cols.push(xPlot)
      cols.push(yPlot)

      c3.generate({
        bindto: '#chart',
        padding: {
          right: 50
        },
        data: {
          x: 'x',
          xFormat: '%Y-%m-%dT%H:%M:%S.%LZ',
          columns: cols
        },
        axis: {
          x: {
            label: {
              text: 'time',
              position: 'outer-middle'
            },
            tick: {
              fit: true,
              format:function (x) {
                  var formatSeconds = d3.time.format("%H:%M:%S")
                  return formatSeconds(new Date(x*1000)); },
              count: 8,
              rotate: 45
            }
          },
          y: {
            label: {
              text: 'queries',
              position: 'outer-middle'
            }
          }
        }
      })

      self.loading = false
      $('#chart').show()
    },
    getActive: function () {
      var self = this
      $.get(apiURL + 'application/active', function (data) {
        self.active = data.active
      })
    },
    setActive: function (state) {
      var self = this
      state = state ? "On" : "Off"
      $.ajax({
        url: apiURL + 'application/active?v=1&state=' + state,
        type: 'PUT',
        success: function (data) {
          self.active = data.active
        }
      })
    },
    pollActive: function () {
      var self = this
      setInterval(self.getActive, 1000)
    },
    toggle_autoupdate: function () {
      var self = this
      if (self.autoUpdateInterval <= 0) {
        self.autoUpdateInterval = 10
      }
      var interval = self.autoUpdateInterval * 1000
      self.autoUpdate = !self.autoUpdate
      if (self.autoUpdate == true) {
        self.autoUpdateId = setInterval(function () {
          self.fetchStats()
        }.bind(self), interval);
      } else {
        if (self.autoUpdateId != 0) {
          clearInterval(self.autoUpdateId)
        }
      }
    }
  }
})
