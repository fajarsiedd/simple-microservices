import { Test, TestingModule } from '@nestjs/testing';
import { ProductsService } from '../services/products.service';
import { RedisService } from '@liaoliaots/nestjs-redis';
import { IProductRepository } from '../interfaces/product-repository.interface';
import { RabbitmqService } from '../../shared/messaging/rabbitmq.service';
import { NotFoundException } from '@nestjs/common';
import { DataSource } from "typeorm";

const mockProductRepository = () => ({
    create: jest.fn(),
    findAll: jest.fn(),
    findOneById: jest.fn(),
    update: jest.fn(),
    remove: jest.fn(),
});

const mockRabbitmqService = () => ({
    publishEvent: jest.fn(),
});

const mockRedisClient = {
    get: jest.fn(),
    set: jest.fn(),
    del: jest.fn(),
};

const mockRedisService = () => ({
    getOrThrow: jest.fn(() => mockRedisClient),
});

const mockQueryBuilder = {
  createQueryBuilder: jest.fn(() => ({
    setLock: jest.fn().mockReturnThis(),
    where: jest.fn().mockReturnThis(),
    getOne: jest.fn(),
  })),
};

const mockManager = {
  getRepository: jest.fn(() => mockQueryBuilder),
  save: jest.fn(),
};

const mockDataSource = {
    transaction: jest.fn((callback) => callback(mockManager)),
};

describe('ProductsService', () => {
    let service: ProductsService;
    let productRepository: ReturnType<typeof mockProductRepository>;
    let rabbitmqService: ReturnType<typeof mockRabbitmqService>;
    let redisClient: typeof mockRedisClient;
    let mockRedisServiceInstance: ReturnType<typeof mockRedisService>;

    beforeEach(async () => {
        const module: TestingModule = await Test.createTestingModule({
            providers: [
                ProductsService,
                {
                    provide: 'IProductRepository',
                    useValue: mockProductRepository(),
                },
                {
                    provide: RabbitmqService,
                    useValue: mockRabbitmqService(),
                },
                {
                    provide: RedisService,
                    useValue: mockRedisService(),
                },
                {
                    provide: DataSource,
                    useValue: mockDataSource,
                },
            ],
        }).compile();

        service = module.get<ProductsService>(ProductsService);
        productRepository = module.get<IProductRepository>('IProductRepository') as any;
        rabbitmqService = module.get<RabbitmqService>(RabbitmqService) as any;
        mockRedisServiceInstance = module.get<RedisService>(RedisService) as any;
        redisClient = mockRedisServiceInstance.getOrThrow();
    });

    it('should be defined', () => {
        expect(service).toBeDefined();
    });
    

    describe('findOne()', () => {
        it('should return a product from cache if it exists', async () => {
            const product = { id: 1, name: 'Basmut', price: 1000 };
            redisClient.get.mockResolvedValue(JSON.stringify(product));

            const result = await service.findOne(1);

            expect(redisClient.get).toHaveBeenCalledWith('products:id:1');
            expect(productRepository.findOneById).not.toHaveBeenCalled();
            expect(result).toEqual(product);
        });

        it('should return a product from the repository and cache it if not in cache', async () => {
            const product = { id: 2, name: 'Buryam', price: 25 };
            redisClient.get.mockResolvedValue(null);
            productRepository.findOneById.mockResolvedValue(product);
            redisClient.set.mockResolvedValue('OK');

            const result = await service.findOne(2);

            expect(redisClient.get).toHaveBeenCalledWith('products:id:2');
            expect(productRepository.findOneById).toHaveBeenCalledWith(2);
            expect(redisClient.set).toHaveBeenCalledWith('products:id:2', JSON.stringify(product));
            expect(result).toEqual(product);
        });

        it('should throw NotFoundException if product does not exist', async () => {
            redisClient.get.mockResolvedValue(null);
            productRepository.findOneById.mockResolvedValue(null);

            await expect(service.findOne(999)).rejects.toThrow(NotFoundException);
        });
    });

    describe('create()', () => {
        it('should create a new product, clear cache, and publish an event', async () => {
            const createDto = { name: 'Nasgor', price: 20000 };
            const newProduct = { id: 1, ...createDto };

            productRepository.create.mockResolvedValue(newProduct);
            redisClient.del.mockResolvedValue(1);

            const result = await service.create(createDto as any);

            expect(productRepository.create).toHaveBeenCalledWith(createDto);
            expect(redisClient.del).toHaveBeenCalledWith('products:id:1');
            expect(rabbitmqService.publishEvent).toHaveBeenCalledWith('product.created', newProduct);
            expect(result).toEqual(newProduct);
        });
    });
});