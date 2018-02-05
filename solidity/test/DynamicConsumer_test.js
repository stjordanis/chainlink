'use strict';

require('./support/helpers.js')

contract('DynamicConsumer', () => {
  let Oracle = artifacts.require("./contracts/Oracle.sol");
  let Consumer = artifacts.require("./test/contracts/DynamicConsumer.sol");
  let jobId = "4c7b7ffb66b344fbaa64995af81e355a";
  let oc, cc;

  beforeEach(async () => {
    oc = await Oracle.new({from: oracleNode});
    cc = await Consumer.new(oc.address, {from: stranger});
  });

  it("has a predictable gas price", async () => {
    let rec = await eth.getTransactionReceipt(cc.transactionHash);
    assert.isBelow(rec.gasUsed, 910000);
  });

  describe("#requestEthereumPrice", () => {
    it("triggers a log event in the Oracle contract", async () => {
      let tx = await cc.requestEthereumPrice("usd");

      let events = await getEvents(oc);
      assert.equal(1, events.length)
      let event = events[0]
      assert.equal(event.args.data, `{"url":"https://etherprice.com/api","path":["recent","usd"]}`)
      assert.equal(web3.toUtf8(event.args.jobId), "someJobId");
    });

    it("has a reasonable gas cost", async () => {
      let tx = await cc.requestEthereumPrice("usd");
      assert.isBelow(tx.receipt.gasUsed, 120000);
    });
  });

  describe("#fulfillData", () => {
    let response = "1,000,000.00";
    let nonce;

    beforeEach(async () => {
      await cc.requestEthereumPrice("usd");
      let event = await getLatestEvent(oc);
      nonce = event.args.nonce
    });

    it("records the data given to it by the oracle", async () => {
      await oc.fulfillData(nonce, response, {from: oracleNode})

      let received = await cc.currentPrice.call();
      assert.equal(web3.toUtf8(received), response);
    });

    context("when the consumer does not recognize the nonce", () => {
      beforeEach(async () => {
        await oc.requestData(jobId, cc.address, functionID("fulfill(uint256,bytes32)"), "");
        let event = await getLatestEvent(oc);
        nonce = event.args.nonce
      });

      it("does not accept the data provided", async () => {
        let tx = await cc.requestEthereumPrice("usd");

        await assertActionThrows(async () => {
          await oc.fulfillData(nonce, response, {from: oracleNode})
        });

        let received = await cc.currentPrice.call();
        assert.equal(web3.toUtf8(received), "");
      });
    });

    context("when called by anyone other than the oracle contract", () => {
      it("does not accept the data provided", async () => {
        await assertActionThrows(async () => {
          await cc.fulfill(nonce, response, {from: oracleNode})
        });

        let received = await cc.currentPrice.call();
        assert.equal(web3.toUtf8(received), "");
      });
    });
  });
});