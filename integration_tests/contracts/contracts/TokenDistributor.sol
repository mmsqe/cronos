// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;
import {ReentrancyGuard} from "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import {IERC20, SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "./IMintedToken.sol";

import "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title TokenDistributor
 * @notice It handles the distribution of MTD token.
 * It auto-adjusts block rewards over a set number of periods.
 */
contract TokenDistributor is Ownable, ReentrancyGuard {
    using SafeERC20 for IERC20;
    using SafeERC20 for IMintedToken;
    uint256 public constant PRECISION_FACTOR = 10**18;
    struct StakingPeriod {
        uint256 rewardPerBlockForStaking;
        uint256 rewardPerBlockForOthers;
        uint256 periodLengthInBlock;
    }

    IMintedToken public immutable mintedToken;

    address public immutable tokenSplitter;
    address public stakingAddr;

    // Number of reward periods
    uint256 public immutable NUMBER_PERIODS;

    // Block number when rewards start
    uint256 public immutable START_BLOCK;

    // Current phase for rewards
    uint256 public currentPhase;

    // Block number when rewards end
    uint256 public endBlock;

    // Block number of the last update
    uint256 public lastRewardBlock;

    // Tokens distributed per block for other purposes (team + treasury + trading rewards)
    uint256 public rewardPerBlockForOthers;

    // Tokens distributed per block for staking
    uint256 public rewardPerBlockForStaking;

    mapping(uint256 => StakingPeriod) public stakingPeriod;

    event NewRewardsPerBlock(
        uint256 indexed currentPhase,
        uint256 startBlock,
        uint256 rewardPerBlockForStaking,
        uint256 rewardPerBlockForOthers
    );
    event SetupStakingAddress(address stakingAddr);

    /**
     * @notice Constructor
     * @param _mintedToken MTD token address
     * @param _tokenSplitter token splitter contract address (for team and trading rewards)
     * @param _startBlock start block for reward program
     * @param _rewardsPerBlockForStaking array of rewards per block for staking
     * @param _rewardsPerBlockForOthers array of rewards per block for other purposes (team + treasury + trading rewards)
     * @param _periodLengthesInBlocks array of period lengthes
     * @param _numberPeriods number of periods with different rewards/lengthes (e.g., if 3 changes --> 4 periods)
     */
    constructor(
        address _mintedToken,
        address _tokenSplitter,
        uint256 _startBlock,
        uint256[] memory _rewardsPerBlockForStaking,
        uint256[] memory _rewardsPerBlockForOthers,
        uint256[] memory _periodLengthesInBlocks,
        uint256 _numberPeriods,
        uint256 slippage
    ) {
        require(_mintedToken != address(0), "Distributor: mintedToken must not be address(0)");
        require(
            (_periodLengthesInBlocks.length == _numberPeriods) &&
                (_rewardsPerBlockForStaking.length == _numberPeriods) &&
                (_rewardsPerBlockForOthers.length == _numberPeriods),
            "Distributor: lengths must match numberPeriods"
        );
        require(_tokenSplitter != address(0), "Distributor: tokenSplitter must not be address(0)");

        // // 1. Operational checks for supply
        // uint256 nonCirculatingSupply = IMintedToken(_mintedToken).SUPPLY_CAP() -
        //     IMintedToken(_mintedToken).totalSupply();

        // uint256 amountTokensToBeMinted;

        // for (uint256 i = 0; i < _numberPeriods; i++) {
        //     amountTokensToBeMinted +=
        //         (_rewardsPerBlockForStaking[i] * _periodLengthesInBlocks[i]) +
        //         (_rewardsPerBlockForOthers[i] * _periodLengthesInBlocks[i]);

        //     stakingPeriod[i] = StakingPeriod({
        //         rewardPerBlockForStaking: _rewardsPerBlockForStaking[i],
        //         rewardPerBlockForOthers: _rewardsPerBlockForOthers[i],
        //         periodLengthInBlock: _periodLengthesInBlocks[i]
        //     });
        // }
        // require(amountTokensToBeMinted <= nonCirculatingSupply, "Distributor: rewards exceeds supply");
        // uint256 residueAmt = nonCirculatingSupply - amountTokensToBeMinted;
        // uint256 rewardSlippage = (residueAmt * 100 * PRECISION_FACTOR) / nonCirculatingSupply;
        // require(rewardSlippage <= slippage, "Distributor: slippage exceeds");
        // // 2. Store values
        // mintedToken = IMintedToken(_mintedToken);
        // tokenSplitter = _tokenSplitter;
        // rewardPerBlockForStaking = _rewardsPerBlockForStaking[0];
        // rewardPerBlockForOthers = _rewardsPerBlockForOthers[0];

        // START_BLOCK = _startBlock;
        // endBlock = _startBlock + _periodLengthesInBlocks[0];

        // NUMBER_PERIODS = _numberPeriods;

        // // Set the lastRewardBlock as the startBlock
        // lastRewardBlock = _startBlock;
    }

    /**
     * @dev updates the staking adddress as a mintedBoost contract address once it is deployed.
     */

    function setupStakingAddress(address _stakingAddr) external onlyOwner {
        require(_stakingAddr != address(0), "invalid address");
        stakingAddr = _stakingAddr;
        emit SetupStakingAddress(stakingAddr);
    }

    /**
     * @notice Update pool rewards
     */
    function updatePool() external nonReentrant {
        _updatePool();
    }

    /**
     * @notice Update reward variables of the pool
     */
    function _updatePool() internal {
        require(stakingAddr != address(0), "staking address not setup");
        if (block.number <= lastRewardBlock) {
            return;
        }
        (uint256 tokenRewardForStaking, uint256 tokenRewardForOthers) = _calculatePendingRewards();
        // mint tokens only if token rewards for staking are not null
        if (tokenRewardForStaking > 0) {
            // It allows protection against potential issues to prevent funds from being locked
            mintedToken.mint(stakingAddr, tokenRewardForStaking);
            mintedToken.mint(tokenSplitter, tokenRewardForOthers);
        }

        // Update last reward block only if it wasn't updated after or at the end block
        if (lastRewardBlock <= endBlock) {
            lastRewardBlock = block.number;
        }
    }

    function _calculatePendingRewards() internal returns (uint256, uint256) {
        if (block.number <= lastRewardBlock) {
            return (0, 0);
        }
        // Calculate multiplier
        uint256 multiplier = _getMultiplier(lastRewardBlock, block.number, endBlock);
        // Calculate rewards for staking and others
        uint256 tokenRewardForStaking = multiplier * rewardPerBlockForStaking;
        uint256 tokenRewardForOthers = multiplier * rewardPerBlockForOthers;

        // Check whether to adjust multipliers and reward per block
        while ((block.number > endBlock) && (currentPhase < (NUMBER_PERIODS - 1))) {
            // Update rewards per block
            _updateRewardsPerBlock(endBlock);

            uint256 previousEndBlock = endBlock;

            // Adjust the end block
            endBlock += stakingPeriod[currentPhase].periodLengthInBlock;

            // Adjust multiplier to cover the missing periods with other lower inflation schedule
            uint256 newMultiplier = _getMultiplier(previousEndBlock, block.number, endBlock);

            // Adjust token rewards
            tokenRewardForStaking += (newMultiplier * rewardPerBlockForStaking);
            tokenRewardForOthers += (newMultiplier * rewardPerBlockForOthers);
        }
        return (tokenRewardForStaking, tokenRewardForOthers);
    }

    function getPendingRewards() external view returns (uint256, uint256) {
        if (block.number <= lastRewardBlock) {
            return (0, 0);
        }
        // shadow state vars to avoid updates
        uint256 tEndBlock = endBlock;
        uint256 tCurrentPhase = currentPhase;
        uint256 tRewardPerBlockForStaking = rewardPerBlockForStaking;
        uint256 tRewardPerBlockForOthers = rewardPerBlockForOthers;
        // Calculate multiplier
        uint256 multiplier = _getMultiplier(lastRewardBlock, block.number, tEndBlock);
        // Calculate rewards for staking and others
        uint256 tokenRewardForStaking = multiplier * tRewardPerBlockForStaking;
        uint256 tokenRewardForOthers = multiplier * tRewardPerBlockForOthers;
        // Check whether to adjust multipliers and reward per block
        while ((block.number > tEndBlock) && (tCurrentPhase < (NUMBER_PERIODS - 1))) {
            // Update rewards per block
            tCurrentPhase++;
            tRewardPerBlockForStaking = stakingPeriod[tCurrentPhase].rewardPerBlockForStaking;
            tRewardPerBlockForOthers = stakingPeriod[tCurrentPhase].rewardPerBlockForOthers;
            uint256 previousEndBlock = tEndBlock;

            // Adjust the end block
            tEndBlock += stakingPeriod[tCurrentPhase].periodLengthInBlock;

            // Adjust multiplier to cover the missing periods with other lower inflation schedule
            uint256 newMultiplier = _getMultiplier(previousEndBlock, block.number, tEndBlock);

            // Adjust token rewards
            tokenRewardForStaking += (newMultiplier * tRewardPerBlockForStaking);
            tokenRewardForOthers += (newMultiplier * tRewardPerBlockForOthers);
        }
        return (tokenRewardForStaking, tokenRewardForOthers);
    }

    function getPendingStakingRewards() external view returns (uint256) {
        if (block.number <= lastRewardBlock) {
            return 0;
        }
        uint256 multiplier = block.number - lastRewardBlock;
        return multiplier * rewardPerBlockForStaking;
    }

    /**
     * @notice Update rewards per block
     * @dev Rewards are halved by 2 (for staking + others)
     */
    function _updateRewardsPerBlock(uint256 _newStartBlock) internal {
        // Update current phase
        currentPhase++;

        // Update rewards per block
        rewardPerBlockForStaking = stakingPeriod[currentPhase].rewardPerBlockForStaking;
        rewardPerBlockForOthers = stakingPeriod[currentPhase].rewardPerBlockForOthers;

        emit NewRewardsPerBlock(
            currentPhase,
            _newStartBlock,
            rewardPerBlockForStaking,
            rewardPerBlockForOthers
        );
    }

    /**
     * @notice Return reward multiplier over the given "from" to "to" block.
     * @param from block to start calculating reward
     * @param to block to finish calculating reward
     * @return the multiplier for the period
     */
    function _getMultiplier(
        uint256 from,
        uint256 to,
        uint256 tEndBlock
    ) internal pure returns (uint256) {
        if (to <= tEndBlock) {
            return to - from;
        } else if (from >= tEndBlock) {
            return 0;
        } else {
            return tEndBlock - from;
        }
    }
}