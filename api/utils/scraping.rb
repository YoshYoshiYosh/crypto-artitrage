require 'net/http'
require 'nokogiri'

url = "https://info.finance.yahoo.co.jp/fx/detail/?code=usdjpy"
uri = URI(url)
html = Net::HTTP.get(uri)

doc = Nokogiri::HTML.parse(html, nil, 'utf-8')
# puts doc
low_nodes = doc.css('dd#USDJPY_detail_low').text.to_f
heigh_nodes = doc.css('dd#USDJPY_detail_high').text.to_f
print ((low_nodes + heigh_nodes)/2).round(4)


